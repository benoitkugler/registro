package espaceperso

import (
	"database/sql"
	"fmt"

	"registro/config"
	"registro/crypto"
	ds "registro/sql/dossiers"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key  crypto.Encrypter
	smtp config.SMTP
	asso config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso) *Controller {
	return &Controller{db, key, smtp, asso}
}

func (ct *Controller) TmpEspaceperso(c echo.Context) error {
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, c.QueryParam("key"))
	if err != nil {
		return err
	}
	return c.String(200, fmt.Sprintf("Inscription validée: dossier %d", id))
}
