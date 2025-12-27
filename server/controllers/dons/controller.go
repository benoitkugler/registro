package dons

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"registro/config"
	"registro/controllers/files"
	"registro/helloasso"
	"registro/logic"
	"registro/logic/search"
	"registro/mails"
	"registro/recufiscal"
	dn "registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	db   *sql.DB
	asso config.Asso
	smtp config.SMTP

	helloasso config.Helloasso // optionnal
}

// HandleDonHelloasso is the webhook trigerred by Helloasso
// See https://dev.helloasso.com/docs/notifications-webhook
func (ct *Controller) HandleDonHelloasso(c echo.Context) error {
	if ct.helloasso.ID == "" {
		return errors.New("HelloAsso is not supported")
	}
	api := helloasso.NewApi(ct.helloasso)
	req := c.Request()
	if req.Body == nil {
		return errors.New("missing Body")
	}
	defer req.Body.Close()

	don, ok, err := api.HandleDonNotification(req.Body)
	if err != nil {
		return err
	}
	if !ok { // just ignore the notification
		return c.NoContent(200)
	}
	_, err = ct.identifieAddDon(don.Don, don.Donateur)
	if err != nil {
		// error on our side
		log.Println("handling HelloAsso don", don, err)
	}

	return c.NoContent(200)
}

func (ct *Controller) identifieAddDon(don dn.Don, donateur pr.Etatcivil) (pr.IdPersonne, error) {
	// try to find and merge the donateur
	personnes, err := search.SelectAllFieldsForSimilaires(ct.db)
	if err != nil {
		return 0, err
	}
	idDonateur, hasExisting := search.Match(personnes, search.NewPatternsSimilarite(donateur))

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		if hasExisting { // merge fields
			existing, err := pr.SelectPersonne(ct.db, idDonateur)
			if err != nil {
				return utils.SQLError(err)
			}
			existing.Etatcivil, _ = search.Merge(donateur, existing.Etatcivil)
			_, err = existing.Update(tx)
			if err != nil {
				return err
			}
		} else { // create a new profil
			donateur, err := pr.Personne{Etatcivil: donateur}.Insert(tx)
			if err != nil {
				return err
			}
			idDonateur = donateur.Id
		}

		don.IdPersonne = idDonateur.Opt()
		_, err = don.Insert(tx)
		return err
	})

	return idDonateur, err
}

func (ct *Controller) SearchPersonnes(c echo.Context) error {
	search := c.QueryParam("search")
	out, err := logic.SelectPersonne(ct.db, search, true)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) SearchOrganismes(c echo.Context) error {
	search := c.QueryParam("search")
	out, err := ct.searchOrganismes(search)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type OrganismeHeader struct {
	Id  dn.IdOrganisme
	Nom string
}

func (ct *Controller) searchOrganismes(pattern string) ([]OrganismeHeader, error) {
	organismes, err := dn.SelectAllOrganismes(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	pattern = search.Normalize(pattern)
	var out []OrganismeHeader
	for _, organisme := range organismes {
		if strings.Contains(search.Normalize(organisme.Nom), pattern) {
			out = append(out, OrganismeHeader{organisme.Id, organisme.Nom})
		}
	}

	slices.SortFunc(out, func(a, b OrganismeHeader) int { return int(a.Id - b.Id) })

	return out, nil
}

type DonExt struct {
	Don      dn.Don
	Donateur string
}

func (ct *Controller) LoadDons(c echo.Context) error {
	out, err := ct.loadDons()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadDons() ([]DonExt, error) {
	dons, err := dn.SelectAllDons(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db)
}

type PersonneD struct {
	Nom    string
	Prenom string
	Sexe   pr.Sexe

	DateNaissance shared.Date

	Mail string

	Adresse    string
	CodePostal string
	Ville      string
	Pays       pr.Pays
}

func (p PersonneD) personne() pr.Personne {
	return pr.Personne{Etatcivil: pr.Etatcivil{
		Nom:           p.Nom,
		Prenom:        p.Prenom,
		Sexe:          p.Sexe,
		DateNaissance: p.DateNaissance,
		Mail:          p.Mail,
		Adresse:       p.Adresse,
		CodePostal:    p.CodePostal,
		Ville:         p.Ville,
		Pays:          p.Pays,
	}}
}

func (ct *Controller) CreatePersonne(c echo.Context) error {
	var args PersonneD
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := args.personne().Insert(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.JSON(200, out)
}

func (ct *Controller) CreateOrganisme(c echo.Context) error {
	var args dn.Organisme
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := args.Insert(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.JSON(200, out)
}

// CreateDon inscrit le don et envoie un mail de remerciement automatique
// (sauf pour les dons Helloasso)
func (ct *Controller) CreateDon(c echo.Context) error {
	var args dn.Don
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createDon(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createDon(don dn.Don) (dn.Don, error) {
	// select the donateur coordinates
	var (
		contact mails.Contact
		mail    string
	)
	if idPersonne := don.IdPersonne; idPersonne.Valid {
		donateur, err := pr.SelectPersonne(ct.db, idPersonne.Id)
		if err != nil {
			return dn.Don{}, utils.SQLError(err)
		}
		contact = mails.NewContact(&donateur)
		mail = donateur.Mail
	} else {
		organisme, err := dn.SelectOrganisme(ct.db, don.IdOrganisme.Id)
		if err != nil {
			return dn.Don{}, utils.SQLError(err)
		}
		contact = mails.Contact{Prenom: organisme.Nom}
		mail = organisme.Mail
	}

	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		don, err = don.Insert(ct.db)
		if err != nil {
			return err
		}
		if don.ModePaiement == ds.Helloasso {
			return nil
		}

		html, err := mails.NotifieDon(ct.asso, contact, don.Montant)
		if err != nil {
			return err
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(mail, "Remerciement", html, nil, nil)
		return err
	})
	return don, err
}

func (ct *Controller) UpdateDon(c echo.Context) error {
	var args dn.Don
	if err := c.Bind(&args); err != nil {
		return err
	}
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.NoContent(200)
}

func (ct *Controller) DeleteDon(c echo.Context) error {
	id, err := utils.QueryParamInt[dn.IdDon](c, "id")
	if err != nil {
		return err
	}
	_, err = dn.DeleteDonById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.NoContent(200)
}

func (ct *Controller) DownloadRecusFiscaux(c echo.Context) error {
	year, err := utils.QueryParamInt[int](c, "year")
	if err != nil {
		return err
	}
	content, err := recufiscal.Generate(ct.db, year)
	if err != nil {
		return err
	}
	mimeType := files.SetBlobHeader(c, content, fmt.Sprintf("Recus fiscaux %d.zip", year))
	return c.Blob(200, mimeType, content)
}
