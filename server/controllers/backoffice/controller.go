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
	evAPI "registro/controllers/events"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key       crypto.Encrypter
	files     fs.FileSystem
	smtp      config.SMTP
	joomeo    config.Joomeo
	helloasso config.Helloasso

	builtins fs.Builtins
}

func NewController(db *sql.DB, key crypto.Encrypter, files fs.FileSystem, smtp config.SMTP, joomeo config.Joomeo, helloasso config.Helloasso) (*Controller, error) {
	out := &Controller{
		db:        db,
		key:       key,
		files:     files,
		smtp:      smtp,
		joomeo:    joomeo,
		helloasso: helloasso,
	}
	var err error
	out.builtins, err = fs.LoadBuiltins(db)
	return out, err
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
// It may be used to setup a dev. token.
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

type dossiersLoader struct {
	dossiers     ds.Dossiers
	participants map[ds.IdDossier]cps.Participants
	personnes    pr.Personnes
	camps        cps.Camps
	events       evAPI.Loader
}

// wrap the SQL error
func newDossierLoader(db ds.DB, ids ...ds.IdDossier) (*dossiersLoader, error) {
	dossiers, err := ds.SelectDossiers(db, ids...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// select the participants and associated people
	links, err := cps.SelectParticipantsByIdDossiers(db, ids...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	participants := links.ByIdDossier()

	personnes, err := pr.SelectPersonnes(db, append(dossiers.IdResponsables(), links.IdPersonnes()...)...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// load the camps
	camps, err := cps.SelectCamps(db, links.IdCamps()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// load the messages
	ld, err := evAPI.NewLoaderFor(db, ids...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	return &dossiersLoader{dossiers, participants, personnes, camps, ld}, nil
}

func (ld *dossiersLoader) data(id ds.IdDossier) dossierData {
	dossier, participants := ld.dossiers[id], ld.participants[id]
	events := ld.events.EventsFor(id)
	return dossierData{dossier, participants, ld.personnes, ld.camps, events}
}

type dossierData struct {
	dossier      ds.Dossier
	participants cps.Participants // exact list
	personnesM   pr.Personnes     // containing at least the reponsable and participants
	camps        cps.Camps        // containing at least the camps for [participants]
	events       evAPI.Events
}

// responsable always comes first
func (ld *dossierData) personnes() (out []pr.Personne) {
	out = append(out, ld.personnesM[ld.dossier.IdResponsable])
	for _, part := range ld.participants {
		out = append(out, ld.personnesM[part.IdPersonne])
	}
	return out
}
