package backoffice

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
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
	_, isFondSoutien := JWTUser(c)
	var args EventsSendMessageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	_, err := ct.sendMessage(c.Request().Host, args, isFondSoutien)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// add a message and send a notification
func (ct *Controller) sendMessage(host string, args EventsSendMessageIn, isFondSoutien bool) (event evs.Event, _ error) {
	dossier, responsable, err := dossierAndResp(ct.db, args.IdDossier)
	if err != nil {
		return evs.Event{}, utils.SQLError(err)
	}
	url := logic.EspacePersoURL(ct.key, host, args.IdDossier)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		origine := evs.Backoffice
		if isFondSoutien {
			origine = evs.FondSoutien
		}
		event, _, err = evs.CreateMessage(tx, args.IdDossier, time.Now(), evs.EventMessage{Contenu: args.Contenu, Origine: origine})
		if err != nil {
			return err
		}
		// notifie le responsable
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
	message, ok := event.Content.(logic.MessageEvt)
	if !ok {
		return errors.New("invalid event Kind (expected Message)")
	}
	if message.Message.Origine != evs.Backoffice {
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
	_, isFondSoutien := JWTUser(c)
	idDossier, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	err = ct.markMessagesSeen(idDossier, isFondSoutien)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) markMessagesSeen(idDossier ds.IdDossier, isFondSoutien bool) error {
	events, err := logic.LoadEventsByDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	var updates evs.EventMessages
	for _, event := range events.UnreadMessagesFor(isFondSoutien) {
		message := event.Content.Message
		if isFondSoutien {
			message.VuFondSoutien = true
		} else {
			message.VuBackoffice = true
		}
		updates = append(updates, message)
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

func (ct *Controller) EventsSendFacture(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	err = ct.sendFacture(c.Request().Host, id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) sendFacture(host string, id ds.IdDossier) error {
	dossier, responsable, err := dossierAndResp(ct.db, id)
	if err != nil {
		return err
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		event, err := evs.Event{IdDossier: id, Kind: evs.Facture, Created: time.Now()}.Insert(tx)
		if err != nil {
			return err
		}
		url := logic.EspacePersoURL(ct.key, host, id, utils.QPInt("idEvent", event.Id))
		// notifie le responsable
		body, err := mails.NotifieFacture(ct.asso, mails.NewContact(&responsable), url)
		if err != nil {
			return err
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(responsable.Mail, "Demande de règlement", body, dossier.CopiesMails, nil)
		if err != nil {
			return err
		}
		return nil
	})
}

func (ct *Controller) EventsSendDocumentsCampPreview(c echo.Context) error {
	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	out, err := ct.previewSendDocumentsCamp(idCamp)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type DossierDocumentsState struct {
	Id            ds.IdDossier
	Responsable   string
	Participants  string
	DocumentsSent bool
}

type SendDocumentsCampPreview struct {
	Dossiers []DossierDocumentsState // with at least one inscrit
}

func (ct *Controller) previewSendDocumentsCamp(idCamp cps.IdCamp) (SendDocumentsCampPreview, error) {
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return SendDocumentsCampPreview{}, err
	}
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers())
	if err != nil {
		return SendDocumentsCampPreview{}, err
	}
	var out SendDocumentsCampPreview
	for id := range dossiers.Dossiers {
		dossier := dossiers.For(id)
		if _, hasInscrit := dossier.CampsInscrits()[idCamp]; !hasInscrit {
			continue
		}
		out.Dossiers = append(out.Dossiers, DossierDocumentsState{
			Id:            id,
			Responsable:   dossier.Responsable().PrenomNOM(),
			Participants:  dossier.ParticipantsLabels(),
			DocumentsSent: dossier.Events.HasSendCampDocuments(idCamp),
		})
	}
	return out, nil
}

type SendDocumentsCampIn struct {
	IdCamp     cps.IdCamp
	IdDossiers []ds.IdDossier
}

func (ct *Controller) EventsSendDocumentsCamp(c echo.Context) error {
	var args SendDocumentsCampIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	it, err := ct.sendDocumentsCamp(c.Request().Host, args)
	if err != nil {
		return err
	}
	return utils.StreamJSON(c.Response(), it)
}

// SendDocumentsCamp is also used by the "directeur" controller
func SendDocumentsCamp(db *sql.DB, key crypto.Encrypter, asso config.Asso, smtp config.SMTP, host string, args SendDocumentsCampIn) (iter.Seq2[SendProgress, error], error) {
	camp, err := cps.SelectCamp(db, args.IdCamp)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers, err := logic.LoadDossiers(db, args.IdDossiers)
	if err != nil {
		return nil, err
	}
	ids := dossiers.Dossiers.IDs() // ensure unicity
	pool, err := mails.NewPool(smtp, asso.MailsSettings, nil)
	if err != nil {
		return nil, err
	}

	return func(yield func(SendProgress, error) bool) {
		defer pool.Close()

		for index, idDossier := range ids {
			dossier := dossiers.For(idDossier)
			responsable := dossier.Responsable()

			err = utils.InTx(db, func(tx *sql.Tx) error {
				event, err := evs.Event{IdDossier: idDossier, Kind: evs.CampDocs, Created: time.Now()}.Insert(tx)
				if err != nil {
					return err
				}
				err = evs.EventCampDocs{IdEvent: event.Id, IdCamp: camp.Id}.Insert(tx)
				if err != nil {
					return err
				}
				url := logic.EspacePersoURL(key, host, idDossier, utils.QPInt("idEvent", event.Id))
				body, err := mails.NotifieDocumentsCamp(asso, mails.NewContact(&responsable), camp.Label(), url)
				if err != nil {
					return err
				}
				err = pool.SendMail(responsable.Mail, fmt.Sprintf("Documents du séjour %s", camp.Label()), body, dossier.Dossier.CopiesMails, nil)
				if err != nil {
					return err
				}
				return nil
			})
			if !yield(SendProgress{Current: index + 1, Total: len(ids)}, err) {
				return
			}
		}
	}, nil
}

func (ct *Controller) sendDocumentsCamp(host string, args SendDocumentsCampIn) (iter.Seq2[SendProgress, error], error) {
	return SendDocumentsCamp(ct.db, ct.key, ct.asso, ct.smtp, host, args)
}

func (ct *Controller) EventsSendSondages(c echo.Context) error {
	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	it, err := ct.sendSondages(c.Request().Host, idCamp)
	if err != nil {
		return err
	}
	return utils.StreamJSON(c.Response(), it)
}

func (ct *Controller) sendSondages(host string, idCamp cps.IdCamp) (iter.Seq2[SendProgress, error], error) {
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return nil, err
	}
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers())
	if err != nil {
		return nil, err
	}
	ids := dossiers.Dossiers.IDs() // ensure unicity
	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return nil, err
	}

	return func(yield func(SendProgress, error) bool) {
		defer pool.Close()

		for index, idDossier := range ids {
			dossier := dossiers.For(idDossier)
			if _, hasInscrit := dossier.CampsInscrits()[idCamp]; !hasInscrit {
				continue
			}
			responsable := dossier.Responsable()
			err = utils.InTx(ct.db, func(tx *sql.Tx) error {
				event, err := evs.Event{IdDossier: idDossier, Kind: evs.Sondage, Created: time.Now()}.Insert(tx)
				if err != nil {
					return err
				}
				err = evs.EventSondage{IdEvent: event.Id, IdCamp: idCamp}.Insert(tx)
				if err != nil {
					return err
				}
				url := logic.EspacePersoURL(ct.key, host, idDossier, utils.QPInt("idEvent", event.Id))
				body, err := mails.NotifieSondage(ct.asso, mails.NewContact(&responsable), camp.Camp.Label(), url)
				if err != nil {
					return err
				}
				err = pool.SendMail(responsable.Mail, fmt.Sprintf("Avis sur le séjour %s", camp.Camp.Label()), body, dossier.Dossier.CopiesMails, nil)
				if err != nil {
					return err
				}
				return nil
			})
			if !yield(SendProgress{Current: index + 1, Total: len(ids)}, err) {
				return
			}
		}
	}, nil
}

type RelancePaiementIn struct {
	IdDossiers []ds.IdDossier
}

func (ct *Controller) EventsSendRelancePaiement(c echo.Context) error {
	var args RelancePaiementIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	it, err := ct.sendRelancePaiement(c.Request().Host, args)
	if err != nil {
		return err
	}
	return utils.StreamJSON(c.Response(), it)
}

type SendProgress struct {
	Current int
	Total   int
}

func (ct *Controller) sendRelancePaiement(host string, args RelancePaiementIn) (iter.Seq2[SendProgress, error], error) {
	dossiers, err := logic.LoadDossiers(ct.db, args.IdDossiers)
	if err != nil {
		return nil, err
	}
	ids := dossiers.Dossiers.IDs() // ensure unicity
	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return nil, err
	}

	return func(yield func(SendProgress, error) bool) {
		defer pool.Close()

		for index, idDossier := range ids {
			dossier := dossiers.For(idDossier)
			responsable := dossier.Responsable()

			err = utils.InTx(ct.db, func(tx *sql.Tx) error {
				event, err := evs.Event{IdDossier: idDossier, Kind: evs.Facture, Created: time.Now()}.Insert(tx)
				if err != nil {
					return err
				}
				url := logic.EspacePersoURL(ct.key, host, idDossier, utils.QPInt("idEvent", event.Id))
				// notifie le responsable
				body, err := mails.NotifieFacture(ct.asso, mails.NewContact(&responsable), url)
				if err != nil {
					return err
				}
				err = pool.SendMail(responsable.Mail, "Demande de règlement", body, dossier.Dossier.CopiesMails, nil)
				if err != nil {
					return err
				}
				return nil
			})
			if !yield(SendProgress{Current: index + 1, Total: len(ids)}, err) {
				return
			}
		}
	}, nil
}
