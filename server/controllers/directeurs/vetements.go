package directeurs

import (
	filesAPI "registro/controllers/files"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) VetementsGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getVetements(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getVetements(id cps.IdCamp) (cps.ListeVetements, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return cps.ListeVetements{}, err
	}
	return camp.Vetements, nil
}

// VetementsUpdate saves the given content.
func (ct *Controller) VetementsUpdate(c echo.Context) error {
	user := JWTUser(c)
	var args cps.ListeVetements
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.updateVetements(user, args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) updateVetements(id cps.IdCamp, args cps.ListeVetements) (string, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return "", utils.SQLError(err)
	}
	// TODO: we could sanitize HTML
	camp.Vetements = args
	_, err = camp.Update(ct.db)
	if err != nil {
		return "", utils.SQLError(err)
	}
	// return a render key
	doc, err := filesAPI.CampDocument(ct.key, camp, filesAPI.ListeVetements)
	if err != nil {
		return "", err
	}

	return doc.Key, nil
}
