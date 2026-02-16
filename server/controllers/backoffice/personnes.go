package backoffice

import (
	"registro/logic"
	"registro/logic/search"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) PersonnesGet(c echo.Context) error {
	search := c.QueryParam("search")
	out, err := logic.SelectPersonne(ct.db, search, false)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) PersonnesLoad(c echo.Context) error {
	id, err := utils.QueryParamInt[pr.IdPersonne](c, "id")
	if err != nil {
		return err
	}
	out, err := pr.SelectPersonne(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.JSON(200, out)
}

func (ct *Controller) PersonnesCreate(c echo.Context) error {
	pe, err := pr.Personne{}.Insert(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	out := search.NewPersonneHeader(pe)
	return c.JSON(200, out)
}

func (ct *Controller) PersonnesUpdate(c echo.Context) error {
	var args pr.Personne
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.updatePersonne(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) updatePersonne(args pr.Personne) (search.PersonneHeader, error) {
	current, err := pr.SelectPersonne(ct.db, args.Id)
	if err != nil {
		return search.PersonneHeader{}, utils.SQLError(err)
	}

	current.Identite = args.Identite
	current.Publicite = args.Publicite
	_, err = current.Update(ct.db)
	if err != nil {
		return search.PersonneHeader{}, utils.SQLError(err)
	}

	return search.NewPersonneHeader(current), nil
}
