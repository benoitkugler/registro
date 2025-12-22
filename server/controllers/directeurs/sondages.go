package directeurs

import (
	"registro/logic"
	cps "registro/sql/camps"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) SondagesGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getSondages(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getSondages(idCamp cps.IdCamp) (logic.CampSondages, error) {
	sondages, err := logic.LoadSondages(ct.db, []cps.IdCamp{idCamp})
	if err != nil {
		return logic.CampSondages{}, err
	}
	return sondages.For(idCamp), nil
}
