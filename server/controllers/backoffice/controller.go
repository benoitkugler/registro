// Package backoffice implémente la logique
// du client central, qui permet notament de :
//   - créer et gérer des séjours
//   - suivre les inscriptions
//   - gérer les équipiers
//   - gérer les dons
package backoffice

import (
	"database/sql"
	"time"

	"registro/config"
	"registro/controllers/logic"
	"registro/crypto"
	"registro/utils"

	"registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key       crypto.Encrypter
	password  string
	files     fs.FileSystem
	smtp      config.SMTP
	asso      config.Asso
	joomeo    config.Joomeo
	helloasso config.Helloasso

	builtins fs.Builtins
}

func NewController(db *sql.DB, key crypto.Encrypter, password string, files fs.FileSystem, smtp config.SMTP, asso config.Asso, joomeo config.Joomeo, helloasso config.Helloasso) (*Controller, error) {
	builtins, err := fs.LoadBuiltins(db)
	if err != nil {
		return nil, err
	}
	return &Controller{
		db,
		key,
		password,
		files,
		smtp,
		asso,
		joomeo,
		helloasso,
		builtins,
	}, nil
}

// customClaims are custom claims extending default ones.
type customClaims struct {
	IsAdmin bool
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
func (ct *Controller) NewToken(isAdmin bool) (string, error) {
	// Set custom claims
	claims := &customClaims{
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(deltaToken)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString(ct.key[:])
}

// JWTUser expects a JWT authentified request, and must
// only be used in routes protected by [Controller.JWTMiddleware] or [Controller.JWTMiddlewareForQuery]
func JWTUser(c echo.Context) (isAdmin bool) {
	meta := c.Get("user").(*jwt.Token).Claims.(*customClaims) // the token is valid here
	return meta.IsAdmin
}

type LogginOut struct {
	IsValid bool
	Token   string
}

// Loggin is called to enter the web app,
// and returns a token if the password is valid.
func (ct *Controller) Loggin(c echo.Context) error {
	password := c.QueryParam("password")

	var out LogginOut
	if ct.password == password {
		token, err := ct.NewToken(false)
		if err != nil {
			return err
		}
		out = LogginOut{IsValid: true, Token: token}
	}

	return c.JSON(200, out)
}

// ---------------------------------- Shared API ----------------------------------

func (ct *Controller) GetCamps(c echo.Context) error {
	out, err := logic.LoadCamps(ct.db)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) GetStructureaides(c echo.Context) error {
	out, err := camps.SelectAllStructureaides(ct.db)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) SelectPersonne(c echo.Context) error {
	search := c.QueryParam("search")
	out, err := logic.SelectPersonne(ct.db, search, true)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func dossierAndResp(db ds.DB, id ds.IdDossier) (ds.Dossier, pr.Personne, error) {
	dossier, err := ds.SelectDossier(db, id)
	if err != nil {
		return ds.Dossier{}, pr.Personne{}, utils.SQLError(err)
	}
	responsable, err := pr.SelectPersonne(db, dossier.IdResponsable)
	if err != nil {
		return ds.Dossier{}, pr.Personne{}, utils.SQLError(err)
	}
	return dossier, responsable, nil
}
