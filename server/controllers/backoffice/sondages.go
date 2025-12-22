package backoffice

import (
	"slices"

	"registro/logic"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) CampsLoadSondages(c echo.Context) error {
	year, err := utils.QueryParamInt[int](c, "year")
	if err != nil {
		return err
	}
	out, err := ct.getSondages(year)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type SondagesOut struct {
	Camps    []cps.Camp
	Sondages map[cps.IdCamp]logic.CampSondages
}

func (ct *Controller) getSondages(year int) (SondagesOut, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return SondagesOut{}, utils.SQLError(err)
	}
	camps.RestrictByYear(year)

	sondages, err := logic.LoadSondages(ct.db, camps.IDs())
	if err != nil {
		return SondagesOut{}, err
	}

	out := SondagesOut{
		Camps:    make([]cps.Camp, 0, len(camps)),
		Sondages: make(map[cps.IdCamp]logic.CampSondages),
	}
	for id, camp := range camps {
		out.Camps = append(out.Camps, camp)
		out.Sondages[id] = sondages.For(id)
	}

	// more recent first
	slices.SortFunc(out.Camps, func(a, b cps.Camp) int { return -a.DateDebut.Time().Compare(b.DateDebut.Time()) })

	return out, nil
}
