package dons

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

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
