package backoffice

import (
	"database/sql"
	"errors"
	"time"

	"registro/controllers/espaceperso"
	"registro/controllers/logic"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// InscriptionsGet returns the [Dossier]s to be validated.
func (ct *Controller) InscriptionsGet(c echo.Context) error {
	out, err := ct.getInscriptions()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getInscriptions() ([]logic.Inscription, error) {
	dossiers, err := ds.SelectAllDossiers(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers.RestrictByValidated(false)

	return logic.LoadInscriptions(ct.db, dossiers.IDs()...)
}

func (ct *Controller) InscriptionsSearchSimilaires(c echo.Context) error {
	id, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	out, err := logic.SearchSimilaires(ct.db, id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type InscriptionIdentifieIn struct {
	IdDossier ds.IdDossier
	Target    logic.IdentTarget
}

// InscriptionsIdentifiePersonne identifie et renvoie l'inscription
// mise à jour
func (ct *Controller) InscriptionsIdentifiePersonne(c echo.Context) error {
	var args InscriptionIdentifieIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := logic.IdentifiePersonne(ct.db, args.Target)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, args.IdDossier)
	if err != nil {
		return err
	}
	out := l[0]

	return c.JSON(200, out)
}

// InscriptionsHintValide renvoie la suggestion automatique
// de statut pour chaque participant du dossier donné.
//
// Voir aussi [InscriptionsValide] pour l'action effective.
func (ct *Controller) InscriptionsHintValide(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	out, err := ct.hintValideInscription(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

var backofficeRights = logic.StatutBypassRights{ProfilInvalide: true, CampComplet: true, Inscrit: true}

func (ct *Controller) hintValideInscription(id ds.IdDossier) (logic.StatutHints, error) {
	loader, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}

	return loader.StatutHints(ct.db, backofficeRights)
}

// InscriptionsValideIn indique le statut des participants
// à appliquer.
type InscriptionsValideIn struct {
	IdDossier ds.IdDossier
	Statuts   map[cps.IdParticipant]cps.StatutParticipant
	SendMail  bool
}

// InscriptionsValide marque l'inscription comme validée, après s'être assuré
// qu'aucune personne impliquée n'est temporaire.
//
// Le statut des participants est mis à jour
// et un mail d'accusé de réception est envoyé.
func (ct *Controller) InscriptionsValide(c echo.Context) error {
	var args InscriptionsValideIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.valideInscription(c.Request().Host, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) valideInscription(host string, args InscriptionsValideIn) error {
	loader, err := logic.LoadDossier(ct.db, args.IdDossier)
	if err != nil {
		return err
	}
	dossier := loader.Dossier

	// on s'assure qu'aucune personne n'est temporaire
	for _, pe := range loader.Personnes() {
		if pe.IsTemp {
			return errors.New("internal error: Personne should not be temporary")
		}
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		var inscrits, attente []mails.Participant // personne a statuer après validation
		for _, pExt := range loader.ParticipantsExt() {
			participant := pExt.Participant
			// côté backoffice : par simplicité, tous les participants
			// doivent être validés
			newStatut, _ := args.Statuts[participant.Id]
			if newStatut == 0 {
				return errors.New("internal error: missing participant in InscriptionsValideIn.Statuts")
			}

			participant.Statut = newStatut
			_, err = participant.Update(tx)
			if err != nil {
				return err
			}

			mPart := mails.Participant{Personne: pExt.Personne.PrenomNOM(), Camp: pExt.Camp.Label()}
			if newStatut == cps.Inscrit {
				inscrits = append(inscrits, mPart)
			} else {
				attente = append(attente, mPart)
			}
		}

		dossier.IsValidated = true
		_, err = dossier.Update(tx)
		if err != nil {
			return err
		}

		// mark the validation ...
		ev, err := evs.Event{IdDossier: dossier.Id, Kind: evs.Validation, Created: time.Now()}.Insert(tx)
		if err != nil {
			return err
		}
		err = evs.EventValidation{IdEvent: ev.Id}.Insert(tx)
		if err != nil {
			return err
		}

		// ... and notify if required
		if args.SendMail {
			resp := loader.Responsable()
			url := espaceperso.URLEspacePerso(ct.key, host, dossier.Id, utils.QP("origin", "validation"))
			html, err := mails.NotifieValidationInscription(ct.asso, mails.NewContact(&resp), url, inscrits, attente, nil)
			if err != nil {
				return err
			}
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(resp.Mail, "Inscription reçue", html, dossier.CopiesMails, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
