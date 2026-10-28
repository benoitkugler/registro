package compta

import (
	"database/sql"
	"encoding/json"
	"time"

	"registro/crypto"
	ca "registro/sql/compta"
	"registro/sql/shared"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	adminPassword string
	key           crypto.Encrypter
}

const admin ca.IdComptaCamp = -1 // gomacro:no-enum

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

func (ct *Controller) AdminURL(host string) string {
	adminKey, err := shared.Encode(ct.key, UserKey{IsAdmin: true})
	if err != nil {
		panic(err)
	}

	url := shared.BuildUrl(host, "compta", map[string]string{"key": adminKey})
	return url
}

// DirecteurURL builds the crypted URL for the given camp.
func DirecteurURL(key logs.Encrypteur, host string, idCamp int64) (string, error) {
	dirKey, err := shared.Encode(key, UserKey{IsAdmin: false, IdCamp: idCamp})
	if err != nil {
		return "", err
	}
	url := shared.BuildUrl(host, "compta", map[string]string{"key": dirKey})
	return url, nil
}

func (ct *Controller) DevKeys() string {
	adminKey, err := shared.Encode(ct.key, UserKey{IsAdmin: true})
	if err != nil {
		panic(err)
	}
	dirKey, err := shared.Encode(ct.key, UserKey{IsAdmin: false, IdCamp: 184})
	if err != nil {
		panic(err)
	}
	creds := struct {
		AdminKey string
		DirKey   string
	}{adminKey, dirKey}
	b, err := json.Marshal(creds)
	if err != nil {
		panic(err)
	}
	return string(b)
}
