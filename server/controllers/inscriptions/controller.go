package inscriptions

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"registro/crypto"
	"registro/sql/camps"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key crypto.Encrypter
}

func NewController(db *sql.DB, key crypto.Encrypter) *Controller {
	return &Controller{db, key}
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
	PreselectedCamp    string // optionnel
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
		if isOver := camp.DateFin().Time().Before(now); isOver {
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
	IdCamp   camps.IdCamp
	PreIdent string // crypted

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
