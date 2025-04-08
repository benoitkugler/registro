package espaceperso

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"registro/config"
	"registro/controllers/logic"
	"registro/crypto"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	"registro/utils"

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
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, c.QueryParam("token"))
	if err != nil {
		return err
	}
	return c.String(200, fmt.Sprintf("Inscription validée: dossier %d", id))
}

func (ct *Controller) Load(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}

	out, err := ct.load(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Data struct {
	Dossier logic.DossierExt
}

func (ct *Controller) load(id ds.IdDossier) (Data, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return Data{}, err
	}

	return Data{
		dossier.Publish(),
	}, nil
}

type SendMessageIn struct {
	Token string

	Message string
}

// SendMessage inscrit un nouveau message, sans notifications
func (ct *Controller) SendMessage(c echo.Context) error {
	var args SendMessageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.sendMessage(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) sendMessage(args SendMessageIn) (logic.Event, error) {
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return logic.Event{}, err
	}
	event, message, err := events.CreateMessage(ct.db, id, time.Now(), args.Message, events.FromEspaceperso, events.OptIdCamp{})
	if err != nil {
		return logic.Event{}, utils.SQLError(err)
	}

	return logic.Event{
		Id:      event.Id,
		Created: event.Created,
		Content: logic.Message{
			Message: message,
		},
	}, nil
}
