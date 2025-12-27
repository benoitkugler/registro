package dons

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"registro/config"
	"registro/controllers/files"
	"registro/crypto"
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

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
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

// customClaims are custom claims extending default ones.
type customClaims struct {
	jwt.RegisteredClaims
}

func (ct *Controller) JWTMiddleware() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey:    ct.key[:],
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(customClaims) },
	}
	return echojwt.WithConfig(config)
}

// expects the token to be in the `token` query parameters
func (ct *Controller) JWTMiddlewareForQuery() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey:    ct.key[:],
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(customClaims) },
		TokenLookup:   "query:token",
	}
	return echojwt.WithConfig(config)
}

const deltaToken = 3 * 24 * time.Hour

// NewToken generate a connection token.
//
// It may also be used to setup a dev. token.
func (ct *Controller) NewToken() (string, error) {
	// Set custom claims
	claims := &customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(deltaToken)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString(ct.key[:])
}

type LogginOut struct {
	IsValid bool
	Token   string
}

// Loggin is called to enter the web app,
// and returns a token if the password is valid.
func (ct *Controller) Loggin(c echo.Context) error {
	password := c.QueryParam("password")
	out, err := ct.loggin(password)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loggin(password string) (LogginOut, error) {
	if password != ct.password {
		return LogginOut{}, nil
	}
	token, err := ct.NewToken()
	if err != nil {
		return LogginOut{}, err
	}
	return LogginOut{IsValid: true, Token: token}, nil
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
	personnes, err := pr.SelectPersonnes(ct.db, dons.IdPersonnes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	organismes, err := dn.SelectOrganismes(ct.db, dons.IdOrganismes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	out := make([]DonExt, 0, len(dons))
	for _, don := range dons {
		var donateur string
		if don.IdPersonne.Valid {
			donateur = personnes[don.IdPersonne.Id].NOMPrenom()
		} else {
			donateur = organismes[don.IdOrganisme.Id].Nom
		}
		out = append(out, DonExt{Don: don, Donateur: donateur})
	}
	// sorting will be done on client
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
