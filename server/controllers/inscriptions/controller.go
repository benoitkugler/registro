package inscriptions

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/mails"
	"registro/sql/camps"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
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

// PathConfirmeInscription est envoyé par mail
const PathConfirmeInscription = "/inscription/confirme"

type Controller struct {
	db *sql.DB

	key  crypto.Encrypter
	smtp config.SMTP
	asso config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso) *Controller {
	return &Controller{db, key, smtp, asso}
}

// LoadData décode la (potentiel) préinscription et renvoie les
// données des séjours.
func (ct *Controller) LoadData(c echo.Context) error {
	preselected := c.QueryParam("preselected")       // optionnel
	preinscription := c.QueryParam("preinscription") // optionnel

	out, err := ct.loadData(preselected, preinscription)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type DataInscription struct {
	Camps              cps.Camps
	InitialInscription Inscription
	PreselectedCamp    string // optionnel, copy of 'preselected' query param
}

func (ct *Controller) loadData(preselected, preinscription string) (DataInscription, error) {
	camps, err := ct.loadCamps()
	if err != nil {
		return DataInscription{}, err
	}

	var initialInscription Inscription
	if preinscription != "" {
		initialInscription, err = ct.decodePreinscription(preinscription)
		if err != nil {
			return DataInscription{}, err
		}
	}

	initialInscription.PartageAdressesOK = true // OK par défaut
	if initialInscription.Responsable.Pays == "" {
		initialInscription.Responsable.Pays = "FR"
	}

	return DataInscription{
		Camps:              camps,
		PreselectedCamp:    preselected,
		InitialInscription: initialInscription,
	}, nil
}

// loadCamps renvoie les camps ouverts aux inscriptions et non terminés
func (ct *Controller) loadCamps() (cps.Camps, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	now := time.Now()
	for id, camp := range camps {
		if isOver := camp.DateFin().Time().Before(now); isOver || !camp.Ouvert {
			delete(camps, id)
		}
	}
	return camps, nil
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

func (insc *Inscription) Check() error {
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
	IdCamp   camps.IdCamp

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

// envoie un mail de demande de confirmation
func (ct *Controller) saveInscription(host string, publicInsc Inscription) (err error) {
	camps, err := ct.loadCamps()
	if err != nil {
		return err
	}

	if err = publicInsc.Check(); err != nil {
		return err
	}

	insc := in.Inscription{
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
		return err
	}

	var participants in.InscriptionParticipants
	for _, publicPart := range publicInsc.Participants {
		if _, isCampValid := camps[publicPart.IdCamp]; !isCampValid {
			return errors.New("invalid IdCamp")
		}
		part := in.InscriptionParticipant{
			IdCamp:        publicPart.IdCamp,
			Nom:           publicPart.Nom,
			Prenom:        publicPart.Prenom,
			DateNaissance: publicPart.DateNaissance,
			Sexe:          publicPart.Sexe,
			Nationnalite:  publicPart.Nationnalite,
		}
		part.PreIdent, err = ct.decodePreIdent(publicPart.PreIdent)
		if err != nil {
			return err
		}
		participants = append(participants, part)
	}

	// enregistre l'inscription sur la base
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		insc, err = insc.Insert(tx)
		if err != nil {
			return err
		}
		for i := range participants {
			participants[i].IdInscription = insc.Id
		}
		err = in.InsertManyInscriptionParticipants(tx, participants...)
		return err
	})
	if err != nil {
		return err
	}

	// envoie un mail de demande de confirmation

	cryptedId := crypto.EncryptID(ct.key, insc.Id)
	urlValide := utils.BuildUrl(host, PathConfirmeInscription, map[string]string{
		"crypted-id": cryptedId,
	})

	html, err := mails.ConfirmeInscription(ct.asso, mails.Contact{Prenom: insc.Responsable.Prenom, Sexe: insc.Responsable.Sexe}, urlValide)
	if err != nil {
		return err
	}
	if err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(insc.Responsable.Mail, "Vérification de l'adresse mail", html, nil, nil); err != nil {
		return err
	}
	return nil
}

func (ct *Controller) decodePreIdent(crypted string) (in.OptIdPersonne, error) {
	if crypted == "" { // pas de pré identification
		return in.OptIdPersonne{}, nil
	}

	id, err := crypto.DecryptID[pr.IdPersonne](ct.key, crypted)
	return in.OptIdPersonne{Id: id, Valid: true}, err
}
