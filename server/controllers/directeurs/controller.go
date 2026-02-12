// Package directeurs implémente la logique
// du portail des directeurs, qui permet notament de :
//   - suivre les inscriptions d'un séjour
//   - gérer les documents et fiches sanitaires
//   - gérer les équipiers d'un séjour
package directeurs

import (
	"database/sql"
	"errors"
	"time"

	"registro/config"
	"registro/controllers/backoffice"
	"registro/crypto"
	"registro/logic"
	"registro/utils"

	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key      crypto.Encrypter
	shortKey crypto.ShortEncrypter

	files  fs.FileSystem
	smtp   config.SMTP
	asso   config.Asso
	immich config.Immich

	password string // global password

	builtins fs.Builtins
}

func NewController(db *sql.DB, encryptKey, password string, files fs.FileSystem, smtp config.SMTP, asso config.Asso, joomeo config.Immich) (*Controller, error) {
	builtins, err := fs.LoadBuiltins(db)
	if err != nil {
		return nil, err
	}
	return &Controller{
		db,
		crypto.NewEncrypter(encryptKey),
		crypto.NewShortEncrypter(encryptKey),
		files,
		smtp,
		asso,
		joomeo,
		password,
		builtins,
	}, nil
}

// customClaims are custom claims extending default ones.
type customClaims struct {
	IdCamp cps.IdCamp
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
// It may be also used to setup a dev. token.
func (ct *Controller) NewToken(idCamp cps.IdCamp) (string, error) {
	// Set custom claims
	claims := &customClaims{
		IdCamp: idCamp,
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
func JWTUser(c echo.Context) (idCamp cps.IdCamp) {
	meta := c.Get("user").(*jwt.Token).Claims.(*customClaims) // the token is valid here
	return meta.IdCamp
}

type LogginOut struct {
	IsValid bool

	Camp      logic.CampItem
	ComptaURL string
	Token     string
}

// Loggin is called to enter the web app,
// and returns a token if the password is valid.
func (ct *Controller) Loggin(c echo.Context) error {
	password := c.QueryParam("password")
	id, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}

	out, err := ct.loggin(c.Request().Host, id, password)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) loggin(host string, id cps.IdCamp, password string) (LogginOut, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return LogginOut{}, utils.SQLError(err)
	}

	// we support 5 authentication modes:
	//	- temporary token, used when redirecting from backoffice to directeur
	//	- cached token
	//	- global password
	//	- directeur password
	//	- camp password

	var args backoffice.BackofficeToDirecteurKey

	isAllowed := false

	// first check for tokens
	if _, ok := crypto.VerifyJWT[customClaims](ct.key, password); ok {
		isAllowed = true
	} else if err = ct.key.DecryptJSON(password, &args); err == nil {
		// check the validity
		if time.Since(args.Time) > 24*time.Hour {
			return LogginOut{}, errors.New("quick access token has expired")
		}
		if args.IdCamp != id {
			return LogginOut{}, errors.New("quick access token idCamp mismatch")
		}
		isAllowed = true
	} else {
		// fetch the directeur
		equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, id)
		if err != nil {
			return LogginOut{}, utils.SQLError(err)
		}
		var directeurPassword string
		directeur, hasDirecteur := equipiers.Directeur()
		if hasDirecteur {
			directeurPassword = ct.shortKey.ShortKey(directeur.IdPersonne)
		}

		isAllowed = password == ct.password ||
			(hasDirecteur && password == directeurPassword) ||
			password == camp.Password
	}

	if isAllowed {
		token, err := ct.NewToken(id)
		if err != nil {
			return LogginOut{}, err
		}
		comptaURL, err := comptaURL(ct.key, host, id)
		if err != nil {
			return LogginOut{}, err
		}
		item := logic.NewCampItem(camp)
		return LogginOut{true, item, comptaURL, token}, nil
	}
	return LogginOut{IsValid: false}, nil
}

// comptaURL builds the crypted URL for the given camp.
func comptaURL(key crypto.Encrypter, host string, idCamp cps.IdCamp) (string, error) {
	type UserKey struct {
		IsAdmin bool
		IdCamp  cps.IdCamp
	}
	dirKey, err := key.EncryptJSON(UserKey{IsAdmin: false, IdCamp: idCamp})
	if err != nil {
		return "", err
	}
	url := utils.BuildUrl(host, "compta", utils.QP("key", dirKey))
	return url, nil
}

// helpers

// wraps error
func (ct *Controller) findDirecteur(id cps.IdCamp) (pr.Personne, bool, error) {
	eqs, err := cps.SelectEquipiersByIdCamps(ct.db, id)
	if err != nil {
		return pr.Personne{}, false, utils.SQLError(err)
	}
	dir, has := eqs.Directeur()
	if !has {
		return pr.Personne{}, false, nil
	}
	directeur, err := pr.SelectPersonne(ct.db, dir.IdPersonne)
	if err != nil {
		return pr.Personne{}, false, utils.SQLError(err)
	}
	return directeur, true, nil
}

// ---------------------------- shared API ----------------------------

func (ct *Controller) GetCamps(c echo.Context) error {
	out, err := logic.LoadCamps(ct.db)
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
