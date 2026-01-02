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
	"registro/crypto"
	"registro/generators/sheets"
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
	db       *sql.DB
	key      crypto.Encrypter
	password string
	asso     config.Asso
	smtp     config.SMTP

	helloasso config.Helloasso // optionnal
}

// [helloasso] is optionnal
func NewController(db *sql.DB, key crypto.Encrypter, password string, asso config.Asso, smtp config.SMTP, helloasso config.Helloasso) *Controller {
	return &Controller{db, key, password, asso, smtp, helloasso}
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
	Montant  string
	Donateur string
}

func (ct *Controller) LoadDons(c echo.Context) error {
	out, err := ct.loadDons()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type YearTotal struct{ Particuliers, Organismes string }

type DonsOut struct {
	Dons       []DonExt
	YearTotals map[int]YearTotal
}

// sort by time
func (ds DonsOut) byYear(year int) []DonExt {
	var out []DonExt
	for _, don := range ds.Dons {
		if y := don.Don.Date.Time().Year(); y == year {
			out = append(out, don)
		}
	}

	slices.SortFunc(out, func(a, b DonExt) int { return a.Don.Date.Time().Compare(b.Don.Date.Time()) })

	return out
}

func (ct *Controller) loadDons() (DonsOut, error) {
	dons, err := dn.SelectAllDons(ct.db)
	if err != nil {
		return DonsOut{}, utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db, dons.IdPersonnes()...)
	if err != nil {
		return DonsOut{}, utils.SQLError(err)
	}
	organismes, err := dn.SelectOrganismes(ct.db, dons.IdOrganismes()...)
	if err != nil {
		return DonsOut{}, utils.SQLError(err)
	}

	totals := map[int][2]ds.MultiCurrencies{}
	list := make([]DonExt, 0, len(dons))
	for _, don := range dons {
		year := don.Date.Time().Year()
		total := totals[year]

		var donateur string
		if don.IdPersonne.Valid {
			donateur = personnes[don.IdPersonne.Id].NOMPrenom()
			total[0].Add(don.Montant)
		} else {
			donateur = organismes[don.IdOrganisme.Id].Nom
			total[1].Add(don.Montant)
		}

		list = append(list, DonExt{don, don.Montant.String(), donateur})
		totals[year] = total
	}
	// sorting will be done on client

	out := DonsOut{Dons: list, YearTotals: make(map[int]YearTotal)}
	for k, v := range totals {
		out.YearTotals[k] = YearTotal{v[0].String(), v[1].String()}
	}
	return out, nil
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

func (ct *Controller) CreatePersonneDonateur(c echo.Context) error {
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

func (ct *Controller) DownloadDonsExcel(c echo.Context) error {
	year, err := utils.QueryParamInt[int](c, "year")
	if err != nil {
		return err
	}
	content, err := ct.exportDonsExcel(year)
	if err != nil {
		return err
	}
	mimeType := files.SetBlobHeader(c, content, fmt.Sprintf("Dons %d.xlsx", year))
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) exportDonsExcel(year int) ([]byte, error) {
	data, err := ct.loadDons()
	if err != nil {
		return nil, err
	}
	dons := data.byYear(year)
	l := make([][]sheets.Cell, len(dons))
	for i, don := range dons {
		l[i] = []sheets.Cell{
			{ValueF: float32(don.Don.Id), NumFormat: sheets.Int},
			{Value: don.Donateur},
			{Value: don.Montant},
			{Value: don.Don.ModePaiement.String()},
			{Value: don.Don.Date.String()},
			{Value: don.Don.Details},
			{Value: don.Don.Affectation},
		}
	}
	return sheets.CreateTable([]string{"ID", "Donateur", "Montant", "Mode", "Date", "Détails", "Affectation"}, l)
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
