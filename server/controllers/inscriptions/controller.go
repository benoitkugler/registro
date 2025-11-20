package inscriptions

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"slices"
	"sort"
	"strings"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/logic"
	"registro/logic/search"
	"registro/mails"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// La procédure d'inscription se déroule en 4 temps :
//	- [LoadCamps] : les camps ouverts sont retournées au client
//	- [InitInscription] : les camps ouverts sont retournées au client,
//	  avec (en option) les données de pré-inscription
//	- [SaveInscription] : le client envoie une demande d'inscription : le serveur l'enregistre et en
//	  envoie une demande de confirmation par email
//	- [ConfirmeInscription] : le lien de confirmation est activé : l'inscription est confirmée et le dossier
//    est créé (l'espace perso est alors accessible)

const (
	// EndpointInscription est envoyé par mail pour les pré-inscriptions
	// et utilisé pour les liens de pré-selection.
	EndpointInscription    = "/inscription"
	PreselectionQueryParam = "preselected"

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

// InitInscription renvoie les paramètres de l'association et
// la valeur initiale de l'inscription, possiblement déterminée par la préinscription
func (ct *Controller) InitInscription(c echo.Context) error {
	preinscription := c.QueryParam(preinscriptionKey) // optionnel

	out, err := ct.initInscription(preinscription)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

// CampExt is a public version of [cps.Camp]
type CampExt struct {
	Id   cps.IdCamp
	Slug string

	Nom         string
	DateDebut   shared.Date
	Duree       int // nombre de jours date et fin inclus
	Lieu        string
	ImageURL    string // affichée sur le formulaire d'inscription
	Description string // affichée sur le formulaire d'inscription
	Navette     cps.OptionNavette
	Places      int // nombre de places prévues pour le séjour
	AgeMin      int // inclusif
	AgeMax      int // inclusif

	Meta cps.Meta

	// Formatted, possibly including several currencies
	Prix string

	// Nom et prénom du directeur et ses adjoints
	Direction string

	WithoutInscription bool // API visible only
	// Indique si les inscriptions sont encore fermées.
	IsClosed bool
	// Indique si le nombre d'inscrits maximum est atteint
	IsComplet bool
}

func newCampExt(camp cps.Camp, taux ds.Taux, direction []pr.Personne, participants cps.Participants) CampExt {
	chunks := make([]string, len(direction))
	for i, p := range direction {
		chunks[i] = p.PrenomN()
	}
	var dir string
	if len(chunks) == 1 {
		dir = chunks[0]
	} else if len(chunks) >= 2 {
		nminus1 := strings.Join(chunks[:len(chunks)-1], ", ")
		dir = nminus1 + " et " + chunks[len(chunks)-1]
	}
	inscrits := 0
	for _, p := range participants {
		if p.Statut == cps.Inscrit {
			inscrits += 1
		}
	}
	return CampExt{
		Id:   camp.Id,
		Slug: camp.Slug(),

		Nom:         camp.Nom,
		DateDebut:   camp.DateDebut,
		Duree:       camp.Duree,
		Lieu:        camp.Lieu,
		ImageURL:    camp.ImageURL,
		Description: camp.Description,
		Navette:     camp.Navette,
		Places:      camp.Places,
		AgeMin:      camp.AgeMin,
		AgeMax:      camp.AgeMax,
		Meta:        camp.Meta,

		Prix: taux.Convertible(camp.Prix).String(),

		Direction: dir,

		WithoutInscription: camp.WithoutInscription,
		IsClosed:           camp.Statut == cps.VisibleFerme,
		IsComplet:          inscrits >= camp.Places,
	}
}

func (ct *Controller) GetCamps(c echo.Context) error {
	_, out, err := ct.LoadCamps()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Data struct {
	InitialInscription Inscription
	Settings           config.ConfigInscription
}

func (ct *Controller) initInscription(preinscription string) (Data, error) {
	var (
		initialInscription Inscription
		err                error
	)
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

	return Data{
		initialInscription,
		ct.asso.ConfigInscription,
	}, nil
}

// LoadCamps renvoie les camps visibles aux inscriptions et non terminés.
//
// This method is used internally and also exposed as a public API,
// for other frontends.
func (ct *Controller) LoadCamps() (cps.Camps, []CampExt, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	for id, camp := range camps {
		if camp.Ext().IsTerminated || camp.Statut == cps.Ferme {
			delete(camps, id)
		}
	}
	tauxs, err := ds.SelectTauxs(ct.db, camps.IdTauxs()...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	link, personnes, err := cps.LoadEquipiersByCamps(ct.db, camps.IDs()...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	equipiers := link.ByIdCamp()

	tmp, err := cps.SelectParticipantsByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	participants := tmp.ByIdCamp()

	list := make([]CampExt, 0, len(camps))
	for _, camp := range camps {
		eqs := equipiers[camp.Id].Direction()
		direction := make([]pr.Personne, len(eqs))
		for i, eq := range eqs {
			direction[i] = personnes[eq.IdPersonne]
		}
		list = append(list, newCampExt(camp, tauxs[camp.IdTaux], direction, participants[camp.Id]))
	}

	slices.SortFunc(list, func(a, b CampExt) int { return strings.Compare(a.Nom, b.Nom) })

	return camps, list, nil
}

// Inscription est la donnée publique correspondant
// à une inscription.
type Inscription struct {
	Responsable in.ResponsableLegal

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
	IdCamp cps.IdCamp

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
	idsParticipants utils.Set[pr.IdPersonne] // participants cumulés
}

// chercheMail renvoie les personnes ayant le mail fourni. Ignore les personnes temporaires.
func (ct *Controller) chercheMail(mail string) (candidatsPreinscription, error) {
	dossiers, responsables, err := logic.LoadByMail(ct.db, mail)
	if err != nil {
		return candidatsPreinscription{}, err
	}

	out := candidatsPreinscription{
		responsables:    make([]pr.Personne, 0, len(responsables)),
		idsParticipants: make(utils.Set[pr.IdPersonne]),
	}
	for _, pers := range responsables {
		out.responsables = append(out.responsables, pers)
	}
	sort.Slice(out.responsables, func(i int, j int) bool {
		return out.responsables[i].NOMPrenom() < out.responsables[j].NOMPrenom()
	})

	for _, dossier := range dossiers.Dossiers {
		loader := dossiers.For(dossier.Id)
		for _, personne := range loader.Personnes()[1:] {
			if personne.IsTemp {
				continue
			}
			out.idsParticipants.Add(personne.Id)
		}

	}

	return out, nil
}

// preinscription code le choix d'un responsable et des participants associés.
// Cet object est crypté et inséré dans un email
type preinscription struct {
	IdResponsable  pr.IdPersonne
	IdParticipants utils.Set[pr.IdPersonne]
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
		out = append(out, mails.RespoWithLink{Lien: template.HTML(lien), NomPrenom: resp.NOMPrenom()})
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

	for _, part := range parts {
		partInsc := newParticipant(part.Etatcivil)
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
	camps, _, err := ct.LoadCamps()
	if err != nil {
		return insc, ps, err
	}

	if err = publicInsc.check(); err != nil {
		return insc, ps, err
	}

	allTaux := make(utils.Set[ds.IdTaux]) // check that all taux are the same

	for _, publicPart := range publicInsc.Participants {
		camp, isCampValid := camps[publicPart.IdCamp]
		if !isCampValid {
			return insc, ps, errors.New("invalid IdCamp (not existing or open)")
		}
		if camp.WithoutInscription {
			return insc, ps, errors.New("invalid IdCamp (camp without inscription)")
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
		ps = append(ps, part)
	}
	if len(allTaux) != 1 {
		return insc, ps, errors.New("internal error: inconsistent taux")
	}
	sharedTaux := allTaux.Keys()[0] // valid thanks to the check above

	insc = in.Inscription{
		IdTaux:             sharedTaux,
		Responsable:        publicInsc.Responsable,
		Message:            publicInsc.Message,
		CopiesMails:        publicInsc.CopiesMails,
		PartageAdressesOK:  publicInsc.PartageAdressesOK,
		DemandeFondSoutien: publicInsc.DemandeFondSoutien,

		DateHeure:          time.Now().Truncate(time.Second),
		ConfirmedAsDossier: dossiers.OptIdDossier{},
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
		html, err := mails.ValidationMailInscription(ct.asso, mails.Contact{Prenom: insc.Responsable.Prenom, Sexe: insc.Responsable.Sexe}, urlValide)
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
	url := logic.EspacePersoURL(ct.key, c.Request().Host, dossier.Id, utils.QP("from-inscription", "true"))
	return c.Redirect(307, url)
}

// ConfirmeInscription transforme l'inscription en dossier,
// et rapproche (automatiquement) les profils.
//
// Si l'inscription a déjà été validée
func ConfirmeInscription(db *sql.DB, id in.IdInscription) (ds.Dossier, error) {
	insc, err := in.SelectInscription(db, id)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}
	participants, err := in.SelectInscriptionParticipantsByIdInscriptions(db, id)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}

	if insc.ConfirmedAsDossier.Valid {
		// Just redirect:
		dossier, err := ds.SelectDossier(db, insc.ConfirmedAsDossier.Id)
		if err != nil {
			return ds.Dossier{}, utils.SQLError(err)
		}
		return dossier, nil
	}

	// on charge l'index une fois pour toutes ...
	index, err := search.SelectAllFieldsForSimilaires(db)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}
	// ... et les camps et groupes
	camps, err := cps.SelectCamps(db, participants.IdCamps()...)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}
	tmp, err := cps.SelectGroupesByIdCamps(db, camps.IDs()...)
	if err != nil {
		return ds.Dossier{}, utils.SQLError(err)
	}
	groupesByCamp := tmp.ByIdCamp()

	var dossier ds.Dossier
	err = utils.InTx(db, func(tx *sql.Tx) error {
		// mise à jour (ou création) des personnes
		responsable := pr.Etatcivil{
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
		}
		responsablePersonne, err := rapprochePersonne(tx, index, responsable, cps.OptIdCamp{})
		if err != nil {
			return err
		}
		// on active automatiquement l'envoi des pub été/hiver pour le responsable
		responsablePersonne.Publicite = pr.Publicite{
			PubEte:   true,
			PubHiver: true,
		}
		responsablePersonne, err = responsablePersonne.Update(tx)
		if err != nil {
			return err
		}

		// on crée le dossier (requis pour les participants)
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

		// on crée maintenant les participants, avec le statut [AStatuer]
		// le calcul du statut précis est repoussé au moment de la validation humaine
		// mais l'éventuel groupe est calculé

		for _, part := range participants {
			pers := pr.Etatcivil{
				Nom:           part.Nom,
				Prenom:        part.Prenom,
				Sexe:          part.Sexe,
				DateNaissance: part.DateNaissance,
				Nationnalite:  part.Nationnalite,
			}
			personne, err := rapprochePersonne(tx, index, pers, part.IdCamp.Opt())
			if err != nil {
				return err
			}

			participant := cps.Participant{
				IdCamp:     part.IdCamp,
				IdPersonne: personne.Id,
				IdDossier:  dossier.Id,
				IdTaux:     insc.IdTaux,
				Statut:     cps.AStatuer,
			}
			participant, err = participant.Insert(tx)
			if err != nil {
				return err
			}

			groupe, hasFound := groupesByCamp[participant.IdCamp].TrouveGroupe(personne.DateNaissance)
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

		// on insert le message du formulaire
		if content := strings.TrimSpace(insc.Message); content != "" {
			created := insc.DateHeure.Add(time.Second) // on s'assure que le message vient après le moment d'inscription
			_, _, err = events.CreateMessage(tx, dossier.Id, created, events.EventMessage{Contenu: content, Origine: events.Espaceperso})
			if err != nil {
				return err
			}
		}

		// tag the inscription as confirmed
		insc.ConfirmedAsDossier = dossier.Id.Opt()
		_, err = insc.Update(tx)
		return err
	})

	return dossier, err
}

// rapprochePersonne effectue un rattachement automatique avec les profils connus :
//   - on cherche un profil correspondant "exactement"
//   - sinon, on crée un nouveau profil
//
// Si le profil existant est déjà présent sur le séjour, on crée un nouveau profil avec
// le statut temporaire, pour indiquer une situation anormale
//
// Si un nouveau profil (non temporaire) est créé, il est aussi ajouté à l'index
func rapprochePersonne(tx *sql.Tx, index pr.Personnes, incomming pr.Etatcivil,
	idCampToCheck cps.OptIdCamp,
) (pr.Personne, error) {
	var target pr.OptIdPersonne
	match, hasMatch := search.Match(index, search.NewPatternsSimilarite(incomming))
	if hasMatch {
		target = match.Opt()
	}

	// on vérifie que la personne n'est pas déjà présente sur le séjour
	markTemp := false
	if target.Valid && idCampToCheck.Valid {
		_, found, err := cps.SelectParticipantByIdCampAndIdPersonne(tx, idCampToCheck.Id, target.Id)
		if err != nil {
			return pr.Personne{}, err
		}
		if found {
			// c'est embêtant: plutôt que de refuser l'inscription,
			// on préfère créer un profil avec le marqueur IsTemp
			target = pr.OptIdPersonne{}
			markTemp = true
		}
	}

	var (
		out pr.Personne
		err error
	)
	if target.Valid {
		// si l'inscription est préidentifiée, on fusionne automatiquement
		// l'inscription avec le profil existant
		out, err = pr.SelectPersonne(tx, target.Id)
		if err != nil {
			return pr.Personne{}, err
		}
		out.Etatcivil, _ = search.Merge(incomming, out.Etatcivil)
		out, err = out.Update(tx)
		if err != nil {
			return pr.Personne{}, err
		}
	} else {
		// sinon, on crée une nouvelle personne
		out = pr.Personne{Etatcivil: incomming, IsTemp: markTemp}
		out, err = out.Insert(tx)
		if err != nil {
			return pr.Personne{}, err
		}
	}

	if !out.IsTemp {
		index[out.Id] = out
	}

	return out, nil
}
