package backoffice

import (
	"database/sql"
	"errors"
	"time"

	"registro/controllers/espaceperso"
	"registro/controllers/logic"
	"registro/mails"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type EventsSendMessageIn struct {
	IdDossier ds.IdDossier
	Contenu   string
}

// EventsSendMessage creates a new message and sends a notification
// to the responsable.
func (ct *Controller) EventsSendMessage(c echo.Context) error {
	var args EventsSendMessageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	_, err := ct.sendMessage(c.Request().Host, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// add a message and send a notification
func (ct *Controller) sendMessage(host string, args EventsSendMessageIn) (event evs.Event, _ error) {
	dossier, responsable, err := dossierAndResp(ct.db, args.IdDossier)
	if err != nil {
		return evs.Event{}, utils.SQLError(err)
	}
	url := espaceperso.URLEspacePerso(ct.key, host, args.IdDossier)

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		event, err = evs.Event{Kind: evs.Message, Created: time.Now(), IdDossier: args.IdDossier}.Insert(tx)
		if err != nil {
			return err
		}
		err = evs.EventMessage{IdEvent: event.Id, Contenu: args.Contenu, Origine: evs.FromBackoffice}.Insert(tx)
		if err != nil {
			return err
		}
		// notifie
		body, err := mails.NotifieMessage(ct.asso, mails.NewContact(&responsable), args.Contenu, url)
		if err != nil {
			return err
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(responsable.Mail, "Nouveau message", body, dossier.CopiesMails, nil)
		if err != nil {
			return err
		}
		return nil
	})
	return event, err
}

func (ct *Controller) EventsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[evs.IdEvent](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteEvent(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteEvent(id evs.IdEvent) error {
	event, err := logic.LoadEvent(ct.db, id)
	if err != nil {
		return nil
	}
	raw := event.Raw()
	// we only allow deleting Message send by the backoffice
	message, ok := event.Content.(logic.Message)
	if !ok {
		return errors.New("invalid event Kind (expected Message)")
	}
	if message.Message.Origine != evs.FromBackoffice {
		return errors.New("invalid Message.Origine (expected FromBackoffice)")
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = evs.DeleteEventMessagesByIdEvents(tx, id)
		if err != nil {
			return err
		}
		_, err = evs.DeleteEventMessageVusByIdEvents(ct.db, id)
		if err != nil {
			return err
		}
		raw.Kind = evs.Supprime
		_, err = raw.Update(tx)
		if err != nil {
			return err
		}
		return nil
	})
}

// EventsMarkMessagesSeen set the messages of the given dossier as seen
// by the backoffice.
func (ct *Controller) EventsMarkMessagesSeen(c echo.Context) error {
	idDossier, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	err = ct.markMessagesSeen(idDossier)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) markMessagesSeen(idDossier ds.IdDossier) error {
	events, err := logic.LoadEventsByDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	var updates evs.EventMessages
	for _, event := range events {
		if message, isMessage := event.Content.(logic.Message); isMessage {
			if message.Message.Origine != evs.FromBackoffice {
				message.Message.VuBackoffice = true
				updates = append(updates, message.Message)
			}
		}
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = evs.DeleteEventMessagesByIdEvents(tx, updates.IdEvents()...)
		if err != nil {
			return err
		}
		err = evs.InsertManyEventMessages(tx, updates...)
		return err
	})
}
