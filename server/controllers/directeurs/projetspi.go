package directeurs

import (
	"database/sql"

	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) ProjetSpiLoad(c echo.Context) error {
	user := JWTUser(c)

	projet, _, err := cps.SelectProjetSpiByIdCamp(ct.db, user)
	if err != nil {
		return utils.SQLError(err)
	}
	projet.IdCamp = user // for empty structs

	return c.JSON(200, projet)
}

func (ct *Controller) ProjetSpiUpdate(c echo.Context) error {
	user := JWTUser(c)
	var args cps.ProjetSpi
	if err := c.Bind(&args); err != nil {
		return err
	}
	args.IdCamp = user
	err := ct.updateProjetSpi(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateProjetSpi(args cps.ProjetSpi) error {
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		err := args.Delete(tx)
		if err != nil {
			return err
		}
		return args.Insert(tx)
	})
}
