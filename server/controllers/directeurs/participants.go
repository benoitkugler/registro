package directeurs

import (
	"registro/controllers/logic"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) ParticipantsGet(c echo.Context) error {
	user := JWTUser(c)

	out, err := ct.getParticipants(user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) getParticipants(id cps.IdCamp) ([]logic.ParticipantExt, error) {
	participants, _, err := logic.LoadParticipants(ct.db, id)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

// ParticipantsUpdate modifie les champs d'un participant.
//
// Seuls les champs Statut, Details et Navette sont pris en compte.
//
// Le statut est modifié sans aucune notification.
func (ct *Controller) ParticipantsUpdate(c echo.Context) error {
	var args cps.Participant
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateParticipant(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateParticipant(args cps.Participant) error {
	current, err := cps.SelectParticipant(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	current.Statut = args.Statut
	current.Details = args.Details
	current.Navette = args.Navette
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}
