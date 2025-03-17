package directeurs

import (
	"registro/controllers/backoffice"
	"registro/controllers/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) InscriptionsGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getInscriptions(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getInscriptions(user cps.IdCamp) ([]logic.Inscription, error) {
	parts, err := cps.SelectParticipantsByIdCamps(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers, err := ds.SelectDossiers(ct.db, parts.IdDossiers()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// restrict to new inscriptions
	dossiers.RestrictByValidated(false)

	return logic.LoadInscriptions(ct.db, dossiers.IDs()...)
}

type InscriptionIdentifieIn = backoffice.InscriptionIdentifieIn

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
