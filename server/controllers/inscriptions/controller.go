package inscriptions

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"registro/config"
	"registro/controllers/search"
	"registro/crypto"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// La procédure d'inscription se déroule en 3 temps :
//	- [LoadData] : les camps ouvertss sont retournées au client,
//	  avec (en option) les données de pré-inscription
//	- [SaveInscription] : le client envoie une demande d'inscription : le serveur l'enregistre et en
//	  envoie une demande de confirmation par email
//	- [ConfirmeInscription] : le lien de confirmation est activé : l'inscription est confirmée et le dossier
//    est créé (l'espace perso est alors accessible)

const (
	// EndpointInscription est envoyé par mail pour les pré-inscriptions
	EndpointInscription = "/inscription"
	// EndpointConfirmeInscription est envoyé par mail
	EndpointConfirmeInscription = "/inscription/confirme"
)

type Controller struct {
	db *sql.DB

	key  crypto.Encrypter
	smtp config.SMTP
	asso config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso) *Controller {
	return &Controller{db, key, smtp, asso}
}

const preinscriptionKey = "preinscription"

// LoadData décode la (potentielle) préinscription et renvoie les
// données des séjours.
func (ct *Controller) LoadData(c echo.Context) error {
	preselected := c.QueryParam("preselected")        // optionnel
	preinscription := c.QueryParam(preinscriptionKey) // optionnel

	out, err := ct.loadData(preselected, preinscription)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

// CampExt is a public version of [cps.Camp]
type CampExt struct {
	Id          cps.IdCamp
	Nom         string
	DateDebut   shared.Date
	Duree       int // nombre de jours date et fin inclus
	Lieu        string
	Description string // Description est affichée sur le formulaire d'inscription
	Navette     cps.Navette
	Places      int // nombre de places prévues pour le séjour
	AgeMin      int // inclusif
	AgeMax      int // inclusif

	// Formatted, possibly including several currencies
	Prix string

	// Nom et prénom du directeur et ses adjoints
	Direction string
}

func newCampExt(camp cps.Camp, taux ds.Taux, direction []pr.Personne) CampExt {
	chunks := make([]string, len(direction))
	for i, p := range direction {
		chunks[i] = p.PrenomNOM()
	}
	var dir string
	if len(chunks) == 1 {
		dir = chunks[0]
	} else if len(chunks) >= 2 {
		nminus1 := strings.Join(chunks[:len(chunks)-1], ", ")
		dir = nminus1 + " et " + chunks[len(chunks)-1]
	}

	return CampExt{
		Id:          camp.Id,
		Nom:         camp.Nom,
		DateDebut:   camp.DateDebut,
		Duree:       camp.Duree,
		Lieu:        camp.Lieu,
		Description: camp.Description,
		Navette:     camp.Navette,
		Places:      camp.Places,
		AgeMin:      camp.AgeMin,
		AgeMax:      camp.AgeMax,

		Prix: taux.Convertible(camp.Prix).String(),

		Direction: dir,
	}
}

type Settings struct {
	// PreselectedCamp is deduced from 'preselected' query param,
	// and is 0 if empty or invalid
	PreselectedCamp cps.IdCamp

	SupportBonsCAF     bool
	SupportANCV        bool
	EmailRetraitMedia  string
	ShowFondSoutien    bool
	ShowCharteConduite bool
}

type Data struct {
	Camps              []CampExt
	InitialInscription Inscription
	Settings           Settings
}

func (ct *Controller) loadData(preselected, preinscription string) (Data, error) {
	camps, tauxs, equipiers, personnes, err := ct.LoadCamps()
	if err != nil {
		return Data{}, err
	}

	list := make([]CampExt, 0, len(camps))
	for _, camp := range camps {
		eqs := equipiers[camp.Id].Direction()
		direction := make([]pr.Personne, len(eqs))
		for i, eq := range eqs {
			direction[i] = personnes[eq.IdPersonne]
		}
		list = append(list, newCampExt(camp, tauxs[camp.IdTaux], direction))
	}

	var initialInscription Inscription
	if preinscription != "" {
		initialInscription, err = ct.decodePreinscription(preinscription)
		if err != nil {
			return Data{}, err
		}
	}
	initialInscription.PartageAdressesOK = true // OK par défaut
	if initialInscription.Responsable.Pays == "" {
		initialInscription.Responsable.Pays = "FR"
	}

	// on error, the idPre will be 0
	idPre, _ := utils.ParseInt[cps.IdCamp](preselected)

	return Data{
		Camps:              list,
		InitialInscription: initialInscription,
		Settings: Settings{
			PreselectedCamp:    idPre,
			SupportBonsCAF:     ct.asso.SupportBonsCAF,
			SupportANCV:        ct.asso.SupportANCV,
			EmailRetraitMedia:  ct.asso.EmailRetraitMedia,
			ShowFondSoutien:    ct.asso.ShowFondSoutien,
			ShowCharteConduite: ct.asso.ShowCharteConduite,
		},
	}, nil
}

// LoadCamps renvoie les camps ouverts aux inscriptions et non terminés
func (ct *Controller) LoadCamps() (cps.Camps, ds.Tauxs, map[cps.IdCamp]cps.Equipiers, pr.Personnes, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, nil, nil, nil, utils.SQLError(err)
	}
	for id, camp := range camps {
		if camp.Ext().IsTerminated || !camp.Ouvert {
			delete(camps, id)
		}
	}
	tauxs, err := ds.SelectTauxs(ct.db, camps.IdTauxs()...)
	if err != nil {
		return nil, nil, nil, nil, utils.SQLError(err)
	}
	link, err := cps.SelectEquipiersByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return nil, nil, nil, nil, utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db, link.IdPersonnes()...)
	if err != nil {
		return nil, nil, nil, nil, utils.SQLError(err)
	}

	return camps, tauxs, link.ByIdCamp(), personnes, nil
}

// Inscription est la donnée publique correspondant
// à une inscription. En particulier, les liens
// de pré-identification sont cryptés.
type Inscription struct {
	Responsable         in.ResponsableLegal
	ResponsablePreIdent string // crypted

	Message            string
	CopiesMails        pr.Mails
	PartageAdressesOK  bool
	DemandeFondSoutien bool

	Participants []Participant
}

func (insc *Inscription) check() error {
	if strings.TrimSpace(insc.Responsable.Nom) == "" {
		return errors.New("missing Nom")
	}
	if strings.TrimSpace(insc.Responsable.Prenom) == "" {
		return errors.New("missing Prenom")
	}
	if insc.Responsable.DateNaissance.Time().IsZero() {
		return errors.New("missing DateNaissance")
	}
	if len(insc.Participants) == 0 {
		return errors.New("missing Participants")
	}
	age := insc.Responsable.DateNaissance.Age(time.Now())
	if age < 18 {
		return errors.New("invalid Age")
	}
	for _, part := range insc.Participants {
		if part.DateNaissance.Time().IsZero() {
			return errors.New("missing Participant.DateNaissance")
		}
	}
	return nil
}

// newResponsableLegal renvoie les champs de la personne
// vus comme le responsable d'une inscription
func newResponsableLegal(r pr.Etatcivil) in.ResponsableLegal {
	return in.ResponsableLegal{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		DateNaissance: r.DateNaissance,
		Sexe:          r.Sexe,
		Mail:          r.Mail,
		Tels:          r.Tels,
		Adresse:       r.Adresse,
		CodePostal:    r.CodePostal,
		Ville:         r.Ville,
		Pays:          r.Pays,
	}
}

type Participant struct {
	PreIdent string // crypted
	IdCamp   cps.IdCamp

	Nom           string
	Prenom        string
	DateNaissance shared.Date
	Sexe          pr.Sexe
	Nationnalite  pr.Nationnalite
}

// newParticipant renvoie la personne comme un
// participant d'une inscription
func newParticipant(r pr.Etatcivil) Participant {
	return Participant{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		Sexe:          r.Sexe,
		DateNaissance: r.DateNaissance,
		Nationnalite:  r.Nationnalite,
	}
}

type SearchHistoryOut struct {
	MailFound bool
}

// SearchHistory analyse l'adresse mail donnée et envoie
// un lien d'inscription rapide aux personnes concernées.
func (ct *Controller) SearchHistory(c echo.Context) error {
	mail := c.QueryParam("mail")
	candidats, err := ct.chercheMail(mail)
	if err != nil {
		return err
	}
	if mailFound := len(candidats.responsables) > 0; !mailFound {
		return c.JSON(200, SearchHistoryOut{MailFound: false})
	}

	targets, err := ct.buildPreinscription(c.Request().Host, candidats)
	if err != nil {
		return err
	}

	html, err := mails.Preinscription(ct.asso, mail, targets)
	if err != nil {
		return err
	}
	if err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(mail, "Inscription rapide", html, nil, nil); err != nil {
		return err
	}

	return c.JSON(200, SearchHistoryOut{MailFound: true})
}

type candidatsPreinscription struct {
	responsables    []pr.Personne
	idsParticipants pr.IdPersonneSet // participants cumulés
}

// chercheMail renvoie les personnes ayant le mail fourni. Ignore les personnes temporaires.
func (ct *Controller) chercheMail(mail string) (out candidatsPreinscription, _ error) {
	mail = strings.TrimSpace(mail)
	if len(mail) <= 1 { // no mail
		return out, nil
	}
	respoPs, err := pr.SelectByMail(ct.db, mail)
	if err != nil {
		return out, utils.SQLError(err)
	}
	respoPs.RemoveTemp()

	ids := pr.IdPersonneSet{}
	for id, pers := range respoPs {
		// ajoute seulement les personnes majeurs, afin d'éviter les
		// confusions dans le cas ou un enfant a la même adresse que ses parents
		if pers.Age() < 18 {
			continue
		}

		ids.Add(id)
		out.responsables = append(out.responsables, pers)
	}
	sort.Slice(out.responsables, func(i int, j int) bool {
		return out.responsables[i].NomPrenom() < out.responsables[j].NomPrenom()
	})

	dossiers, err := ds.SelectDossiersByIdResponsables(ct.db, ids.Keys()...)
	if err != nil {
		return out, utils.SQLError(err)
	}
	participants, err := cps.SelectParticipantsByIdDossiers(ct.db, dossiers.IDs()...)
	if err != nil {
		return out, utils.SQLError(err)
	}
	partPs, err := pr.SelectPersonnes(ct.db, participants.IdPersonnes()...)
	if err != nil {
		return out, utils.SQLError(err)
	}
	partPs.RemoveTemp()
	out.idsParticipants = pr.NewIdPersonneSetFrom(partPs.IDs())
	return out, nil
}

// preinscription code le choix d'un responsable et des participants associés.
// Cet object est crypté et inséré dans un email
type preinscription struct {
	IdResponsable  pr.IdPersonne
	IdParticipants pr.IdPersonneSet
}

func (ct Controller) buildPreinscription(host string, cd candidatsPreinscription) ([]mails.RespoWithLink, error) {
	var out []mails.RespoWithLink
	for _, resp := range cd.responsables {
		t := preinscription{IdResponsable: resp.Id, IdParticipants: cd.idsParticipants}
		crypted, err := ct.key.EncryptJSON(t)
		if err != nil {
			return nil, err
		}
		lien := utils.BuildUrl(host, EndpointInscription, utils.QP(preinscriptionKey, crypted))
		out = append(out, mails.RespoWithLink{Lien: template.HTML(lien), NomPrenom: resp.NomPrenom()})
	}
	return out, nil
}

// decodePreinscription décode le lien du mail et forme une inscription pré-remplie.
func (ct *Controller) decodePreinscription(crypted string) (insc Inscription, _ error) {
	var pre preinscription

	if err := ct.key.DecryptJSON(crypted, &pre); err != nil {
		return insc, fmt.Errorf("invalid preinscription link: %s", err)
	}

	respo, err := pr.SelectPersonne(ct.db, pre.IdResponsable)
	if err != nil {
		return insc, utils.SQLError(err)
	}
	parts, err := pr.SelectPersonnes(ct.db, pre.IdParticipants.Keys()...)
	if err != nil {
		return insc, utils.SQLError(err)
	}

	insc.Responsable = newResponsableLegal(respo.Etatcivil)
	insc.ResponsablePreIdent = crypto.EncryptID(ct.key, respo.Id)

	for _, part := range parts {
		partInsc := newParticipant(part.Etatcivil)
		partInsc.PreIdent = crypto.EncryptID(ct.key, part.Id)

		insc.Participants = append(insc.Participants, partInsc)
	}
	return insc, nil
}

// SaveInscription vérifie et sauvegarde l'inscription, et
// demande une confirmation par mail.
func (ct *Controller) SaveInscription(c echo.Context) error {
	var args Inscription
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := ct.saveInscription(c.Request().Host, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) BuildInscription(publicInsc Inscription) (insc in.Inscription, ps in.InscriptionParticipants, _ error) {
	camps, _, _, _, err := ct.LoadCamps()
	if err != nil {
		return insc, ps, err
	}

	if err = publicInsc.check(); err != nil {
		return insc, ps, err
	}

	allTaux := ds.IdTauxSet{} // check that all taux are the same

	for _, publicPart := range publicInsc.Participants {
		camp, isCampValid := camps[publicPart.IdCamp]
		if !isCampValid {
			return insc, ps, errors.New("invalid IdCamp")
		}
		allTaux.Add(camp.IdTaux)

		part := in.InscriptionParticipant{
			IdCamp:        publicPart.IdCamp,
			IdTaux:        camp.IdTaux,
			Nom:           publicPart.Nom,
			Prenom:        publicPart.Prenom,
			DateNaissance: publicPart.DateNaissance,
			Sexe:          publicPart.Sexe,
			Nationnalite:  publicPart.Nationnalite,
		}
		part.PreIdent, err = ct.decodePreIdent(publicPart.PreIdent)
		if err != nil {
			return insc, ps, err
		}
		ps = append(ps, part)
	}
	if len(allTaux) != 1 {
		return insc, ps, errors.New("internal error: inconsistent taux")
	}
	sharedTaux := allTaux.Keys()[0] // valid thanks to the check above

	// check that, if the participant is pre-identified,
	// the personne is not already present in the camp,
	// so that the error does not trigger in [confirmeInscription]
	tmp, err := cps.SelectParticipantsByIdCamps(ct.db, ps.IdCamps()...)
	if err != nil {
		return insc, ps, utils.SQLError(err)
	}
	currentParticipants := tmp.ByIdCamp()
	for _, part := range ps {
		if !part.PreIdent.Valid {
			continue // nothing to check: a new profil will be created
		}
		idPersonne := part.PreIdent.Id
		campParts := currentParticipants[part.IdCamp]
		if len(campParts.ByIdPersonne()[idPersonne]) != 0 {
			return insc, ps, fmt.Errorf("%s est déjà inscrit sur le camp %s", part.Prenom, camps[part.IdCamp].Label())
		}
	}

	insc = in.Inscription{
		IdTaux:             sharedTaux,
		Responsable:        publicInsc.Responsable,
		Message:            publicInsc.Message,
		CopiesMails:        publicInsc.CopiesMails,
		PartageAdressesOK:  publicInsc.PartageAdressesOK,
		DemandeFondSoutien: publicInsc.DemandeFondSoutien,

		DateHeure:   time.Now(),
		IsConfirmed: false,
	}
	insc.ResponsablePreIdent, err = ct.decodePreIdent(publicInsc.ResponsablePreIdent)
	if err != nil {
		return insc, ps, err
	}

	return insc, ps, nil
}

// envoie un mail de demande de confirmation
func (ct *Controller) saveInscription(host string, publicInsc Inscription) (err error) {
	insc, participants, err := ct.BuildInscription(publicInsc)
	if err != nil {
		return err
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// enregistre l'inscription sur la base
		insc, err = in.Create(tx, insc, participants)
		if err != nil {
			return err
		}

		// envoie un mail de demande de confirmation
		cryptedId := crypto.EncryptID(ct.key, insc.Id)
		urlValide := utils.BuildUrl(host, EndpointConfirmeInscription, utils.QP(queryParamIdInscription, cryptedId))
		html, err := mails.ConfirmeInscription(ct.asso, mails.Contact{Prenom: insc.Responsable.Prenom, Sexe: insc.Responsable.Sexe}, urlValide)
		if err != nil {
			return err
		}
		if err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(insc.Responsable.Mail, "Vérification de l'adresse mail", html, nil, nil); err != nil {
			return err
		}
		return nil
	})

	return err
}

const queryParamIdInscription = "insc-token"

func (ct *Controller) decodePreIdent(crypted string) (pr.OptIdPersonne, error) {
	if crypted == "" { // pas de pré identification
		return pr.OptIdPersonne{}, nil
	}

	id, err := crypto.DecryptID[pr.IdPersonne](ct.key, crypted)
	return id.Opt(), err
}

const EndpointEspacePerso = "espace-perso"

func URLEspacePerso(key crypto.Encrypter, host string, dossier ds.IdDossier, queryParams ...utils.QParam) string {
	crypted := crypto.EncryptID(key, dossier)
	queryParams = append(queryParams, utils.QP("key", crypted))
	return utils.BuildUrl(host, EndpointEspacePerso, queryParams...)
}

// ConfirmeInscription valide l'inscription et crée le [Dossier] associé,
// redirigeant ensuite vers l'espace perso.
func (ct *Controller) ConfirmeInscription(c echo.Context) error {
	idCrypted := c.QueryParam(queryParamIdInscription)
	id, err := crypto.DecryptID[in.IdInscription](ct.key, idCrypted)
	if err != nil {
		return err
	}
	dossier, err := ConfirmeInscription(ct.db, id)
	if err != nil {
		return err
	}
	url := URLEspacePerso(ct.key, c.Request().Host, dossier.Id, utils.QP("from-inscription", "true"))
	return c.Redirect(307, url)
}

// transforme l'inscription en dossier et l'enregistre,
// renvoyant le dossier créé
func ConfirmeInscription(db *sql.DB, id in.IdInscription) (ds.Dossier, error) {
	insc, err := in.SelectInscription(db, id)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}
	participants, err := in.SelectInscriptionParticipantsByIdInscriptions(db, id)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}

	if insc.IsConfirmed {
		return ds.Dossier{}, errors.New("inscription déjà confirmé")
	}

	type identifiedPersonne struct {
		personne pr.Personne
		preIdent pr.OptIdPersonne
	}

	var (
		// responsable et participants
		allPers    []identifiedPersonne
		allPersIDs = pr.IdPersonneSet{}
	)

	responsable := pr.Personne{
		Etatcivil: pr.Etatcivil{
			Nom:           insc.Responsable.Nom,
			Prenom:        insc.Responsable.Prenom,
			DateNaissance: insc.Responsable.DateNaissance,
			Sexe:          insc.Responsable.Sexe,
			Mail:          insc.Responsable.Mail,
			Tels:          insc.Responsable.Tels,
			Adresse:       insc.Responsable.Adresse,
			CodePostal:    insc.Responsable.CodePostal,
			Ville:         insc.Responsable.Ville,
			Pays:          insc.Responsable.Pays,
		},
	}

	allPers = append(allPers, identifiedPersonne{responsable, insc.ResponsablePreIdent})
	allPersIDs.Add(insc.ResponsablePreIdent.Id) // maybe 0, which will be ignored
	for _, part := range participants {
		pers := pr.Personne{Etatcivil: pr.Etatcivil{
			Nom:           part.Nom,
			Prenom:        part.Prenom,
			Sexe:          part.Sexe,
			DateNaissance: part.DateNaissance,
			Nationnalite:  part.Nationnalite,
		}}
		allPers = append(allPers, identifiedPersonne{pers, part.PreIdent})
		allPersIDs.Add(part.PreIdent.Id) // maybe 0, which will be ignored
	}

	var dossier ds.Dossier
	err = utils.InTx(db, func(tx *sql.Tx) error {
		// on charge les personnes pour la comparaison ...
		personnes, err := pr.SelectPersonnes(tx, allPersIDs.Keys()...)
		if err != nil {
			return err
		}
		// ... et les camps et groupes
		camps, err := cps.SelectCamps(tx, participants.IdCamps()...)
		if err != nil {
			return err
		}
		tmp, err := cps.SelectGroupesByIdCamps(tx, camps.IDs()...)
		if err != nil {
			return err
		}
		groupes := tmp.ByIdCamp()

		// mise à jour (ou création) des personnes
		for i, inc := range allPers {
			if existante, exists := personnes[inc.preIdent.Id]; inc.preIdent.Valid && exists {
				// si l'inscription est préidentifiée, on fusionne automatiquement
				// l'inscription avec le profil
				existante.Etatcivil, _ = search.Merge(inc.personne.Etatcivil, existante.Etatcivil)
				allPers[i].personne, err = existante.Update(tx)
			} else {
				// sinon, on crée une nouvelle personne temporaire
				inc.personne.IsTemp = true
				allPers[i].personne, err = inc.personne.Insert(tx)
			}
			if err != nil {
				return err
			}
		}

		responsablePersonne := allPers[0].personne // le responsable est en premier
		participantPersonnes := allPers[1:]

		// on active automatiquement l'envoi des pub été/ hiver pour le responsable
		responsablePersonne.Publicite = pr.Publicite{
			PubEte:   true,
			PubHiver: true,
		}
		responsablePersonne, err = responsablePersonne.Update(tx)
		if err != nil {
			return err
		}

		dossier = ds.Dossier{
			IdTaux:             insc.IdTaux,
			IdResponsable:      responsablePersonne.Id,
			CopiesMails:        insc.CopiesMails,
			PartageAdressesOK:  insc.PartageAdressesOK,
			DemandeFondSoutien: insc.DemandeFondSoutien,
			MomentInscription:  insc.DateHeure,
			IsValidated:        false,
		}
		dossier, err = dossier.Insert(tx)
		if err != nil {
			return err
		}

		// on insert le message du formulaire
		if content := strings.TrimSpace(insc.Message); content != "" {
			event := events.Event{
				IdDossier: dossier.Id,
				Kind:      events.Message,
				Created:   time.Now().Add(time.Millisecond), // on s'assure que le message vient après le moment d'inscription
			}
			event, err = event.Insert(tx)
			if err != nil {
				return err
			}
			err = events.EventMessage{
				IdEvent: event.Id,
				Contenu: content, Origine: events.FromEspaceperso,
			}.Insert(tx)
			if err != nil {
				return err
			}
		}

		// on crée maintenant les participants, avec le statut [AStatuer]
		// le calcul du statut précis est repoussé au moment de la validation humaine
		// mais l'éventuel groupe est calculé

		for i, inscPart := range participants {
			personne := participantPersonnes[i].personne // personne créée ou mise à jour
			participant := cps.Participant{
				IdCamp:     inscPart.IdCamp,
				IdPersonne: personne.Id,
				IdDossier:  dossier.Id,
				IdTaux:     insc.IdTaux,
				Statut:     cps.AStatuer,
			}
			participant, err = participant.Insert(tx)
			if err != nil {
				return err
			}

			groupe, hasFound := groupes[participant.IdCamp].TrouveGroupe(personne.DateNaissance)
			if hasFound {
				// on ajoute automatiquement le nouveau participant au groupe
				err = cps.GroupeParticipant{
					IdGroupe: groupe.Id, IdCamp: groupe.IdCamp,
					IdParticipant: participant.Id,
				}.Insert(tx)
				if err != nil {
					return err
				}
			}
		}

		// tag the inscription as Confirmed
		insc.IsConfirmed = true
		_, err = insc.Update(tx)
		return err
	})

	return dossier, err
}
