package backoffice

import (
	"iter"
	"slices"

	inAPI "registro/controllers/inscriptions"
	"registro/mails"
	cps "registro/sql/camps"
	in "registro/sql/inscriptions"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// InscriptionsGetPending returns the [Inscription] waiting for
// the mail adress to be validated. It is usually a "user" error,
// requiring manual handling by the admin.
func (ct *Controller) InscriptionsGetPending(c echo.Context) error {
	out, err := ct.getPendingInscriptions()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type PendingInscriptionsOut struct {
	Inscriptions []PendingInscription
	Camps        cps.Camps
}

type PendingInscription struct {
	Inscription  in.Inscription
	Participants in.InscriptionParticipants
}

func (ct *Controller) getPendingInscriptions() (PendingInscriptionsOut, error) {
	inscriptions, err := in.SelectAllInscriptions(ct.db)
	if err != nil {
		return PendingInscriptionsOut{}, utils.SQLError(err)
	}
	inscriptions.RestrictByConfirmed(false)

	tmp, err := in.SelectInscriptionParticipantsByIdInscriptions(ct.db, inscriptions.IDs()...)
	if err != nil {
		return PendingInscriptionsOut{}, utils.SQLError(err)
	}
	participants := tmp.ByIdInscription()
	camps, err := cps.SelectCamps(ct.db, tmp.IdCamps()...)
	if err != nil {
		return PendingInscriptionsOut{}, utils.SQLError(err)
	}

	out := PendingInscriptionsOut{Camps: camps}
	for _, insc := range inscriptions {
		out.Inscriptions = append(out.Inscriptions, PendingInscription{insc, participants[insc.Id]})
	}
	slices.SortFunc(out.Inscriptions, func(a, b PendingInscription) int { return a.Inscription.DateHeure.Compare(b.Inscription.DateHeure) })

	return out, nil
}

func (ct *Controller) InscriptionsDeletePending(c echo.Context) error {
	id, err := utils.QueryParamInt[in.IdInscription](c, "id")
	if err != nil {
		return err
	}
	err = ct.deletePendingInscription(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deletePendingInscription(id in.IdInscription) error {
	_, err := in.DeleteInscriptionById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

type UpdatePendingInscriptionIn struct {
	Id   in.IdInscription
	Mail string
}

func (ct *Controller) InscriptionsUpdatePending(c echo.Context) error {
	var args UpdatePendingInscriptionIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updatePendingInscription(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updatePendingInscription(args UpdatePendingInscriptionIn) error {
	inscription, err := in.SelectInscription(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	inscription.Responsable.Mail = args.Mail
	_, err = inscription.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

type RelancePendingInscriptionsIn struct {
	Ids []in.IdInscription
}

func (ct *Controller) InscriptionsRelancePending(c echo.Context) error {
	var args RelancePendingInscriptionsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	it, err := ct.relancePendingInscriptions(c.Request().Host, args)
	if err != nil {
		return err
	}
	return utils.StreamJSON(c.Response(), it)
}

func (ct *Controller) relancePendingInscriptions(host string, args RelancePendingInscriptionsIn) (iter.Seq2[SendProgress, error], error) {
	inscriptions, err := in.SelectInscriptions(ct.db, args.Ids...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return nil, err
	}

	return func(yield func(SendProgress, error) bool) {
		defer pool.Close()

		for index, idInscription := range args.Ids {
			insc := inscriptions[idInscription]
			err := inAPI.SendValidationMail(ct.asso, ct.key, pool, host, insc)
			if !yield(SendProgress{Current: index + 1, Total: len(args.Ids)}, err) {
				return
			}
		}
	}, nil
}
