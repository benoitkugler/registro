package compta

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	pe "registro/sql/personnes"
	"registro/utils"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key crypto.Encrypter

	adminPassword         string
	adminNotificationMail string
	files                 filesAPI

	smtp config.SMTP
	asso config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, compta config.Compta, smtp config.SMTP, asso config.Asso) (*Controller, error) {
	files := filesAPI{googleDriveAPI{compta.GoogleDriveKeyFile, compta.GoogleDriveFolderId}, fileSystemAPI{compta.FilesDir}}

	err := files.googleDrive.check()
	if err != nil {
		return nil, err
	}

	return &Controller{db, key, compta.AdminPassword, compta.AdminMail, files, smtp, asso}, nil
}

const userAdmin ca.IdComptaCamp = -1 // gomacro:no-enum

// UserMeta are custom claims extending default ones.
type UserMeta struct {
	User ca.IdComptaCamp // -1 for admin
	jwt.RegisteredClaims
}

func (ct *Controller) JWTMiddleware() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey:    ct.key[:],
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(UserMeta) },
	}
	return echojwt.WithConfig(config)
}

// expects the token to be in the `token` query parameters
func (ct *Controller) JWTMiddlewareForQuery() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey:    ct.key[:],
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(UserMeta) },
		TokenLookup:   "query:token",
	}
	return echojwt.WithConfig(config)
}

const deltaToken = 3 * 24 * time.Hour

func (ct *Controller) newToken(user ca.IdComptaCamp) (string, error) {
	// Set custom claims
	claims := &UserMeta{
		User: user,
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
// only be used in routes protected by `JWTMiddleware`
func JWTUser(c echo.Context) ca.IdComptaCamp {
	meta := c.Get("user").(*jwt.Token).Claims.(*UserMeta) // the token is valid here
	return meta.User
}

// DirecteurURL builds the crypted URL for the given camp.
func DirecteurURL(key crypto.Encrypter, host string, idCamp cps.IdCamp) string {
	dirKey := crypto.EncryptID(key, idCamp)
	url := utils.BuildUrl(host, "compta", utils.QP("directeur-key", dirKey))
	return url
}

func directUploadDirecteurURL(key crypto.Encrypter, host string, idCamp cps.IdCamp, idCompta ca.IdComptaCamp) string {
	dirKey := crypto.EncryptID(key, idCamp)
	url := utils.BuildUrl(host, "/api/v1/compta/records/upload-query", utils.QP("directeur-key", dirKey), utils.QPInt("idCompta", idCompta))
	return url
}

type LogginOut struct {
	IsValid bool

	Token    string
	IdCompta ca.IdComptaCamp
}

// Loggin parse the user key, detecting if the user is an admin
// or a directeur.
// It also accept an already emitted token, checking its validity.
func (ct *Controller) Loggin(c echo.Context) error {
	isAdmin := utils.QueryParamBool(c, "isAdmin")
	key := c.QueryParam("key")

	out, err := ct.loggin(isAdmin, key)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) loggin(isAdmin bool, key string) (token LogginOut, err error) {
	// check for cached tokens
	meta, isValid := crypto.VerifyJWT[UserMeta](ct.key, key)
	if isValid {
		// ensure user match
		if (meta.User == userAdmin) != isAdmin {
			return LogginOut{}, errAccess
		}
		// renew the token
		tk, err := ct.newToken(meta.User)
		if err != nil {
			return LogginOut{}, err
		}
		return LogginOut{IsValid: true, Token: tk, IdCompta: meta.User}, nil
	}

	if isAdmin {
		if key == ct.adminPassword {
			tk, err := ct.newToken(userAdmin)
			if err != nil {
				return LogginOut{}, err
			}
			return LogginOut{IsValid: true, Token: tk, IdCompta: userAdmin}, nil
		} else {
			return LogginOut{IsValid: false}, nil
		}
	}

	idCamp, err := crypto.DecryptID[cps.IdCamp](ct.key, key)
	if err != nil {
		return LogginOut{}, errors.New("Clé de connection invalide")
	}

	compta, err := ct.loadOrCreateCompta(idCamp)
	if err != nil {
		return LogginOut{}, err
	}
	tk, err := ct.newToken(compta.Id)
	if err != nil {
		return LogginOut{}, err
	}
	return LogginOut{IsValid: true, Token: tk, IdCompta: compta.Id}, nil
}

func (ct *Controller) loadOrCreateCompta(idCamp cps.IdCamp) (ca.ComptaCamp, error) {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return ca.ComptaCamp{}, err
	}

	co, found, err := ca.SelectComptaCampByIdCamp(ct.db, camp.Id)
	if err != nil {
		return ca.ComptaCamp{}, err
	}

	if !found { // create a new compta entry for this camp
		co, err = ca.ComptaCamp{IdCamp: camp.Id, FileStatus: ca.FileSystem}.Insert(ct.db)
		if err != nil {
			return ca.ComptaCamp{}, err
		}
	}

	return co, nil
}

func director(db ca.DB, compta ca.ComptaCamp) (cps.Camp, string, bool, error) {
	camp, err := cps.SelectCamp(db, compta.IdCamp)
	if err != nil {
		return cps.Camp{}, "", false, fmt.Errorf("Erreur SQL: %w", err)
	}
	equipiers, err := cps.SelectEquipiersByIdCamps(db, camp.Id)
	if err != nil {
		return cps.Camp{}, "", false, fmt.Errorf("Erreur SQL: %w", err)
	}
	directorE, hasDirector := equipiers.Directeur()
	var director pe.Personne
	if hasDirector {
		director, err = pe.SelectPersonne(db, directorE.IdPersonne)
		if err != nil {
			return cps.Camp{}, "", false, fmt.Errorf("Erreur SQL: %w", err)
		}
		return camp, string(director.Mail), true, nil
	}
	return camp, "", false, nil
}
