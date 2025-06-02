package directeurs

import (
	"fmt"

	filesAPI "registro/controllers/files"
	"registro/generators/pdfcreator"
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
	err := ct.updateVetements(user, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateVetements(id cps.IdCamp, args cps.ListeVetements) error {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	// TODO: we could sanitize HTML
	camp.Vetements = args
	_, err = camp.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) VetementsRender(c echo.Context) error {
	user := JWTUser(c)
	content, filename, err := ct.renderListeVetements(user)
	if err != nil {
		return err
	}
	mimeType := filesAPI.SetBlobHeader(c, content, filename)
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) renderListeVetements(id cps.IdCamp) ([]byte, string, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	content, err := pdfcreator.CreateListeVetements(ct.asso, camp.Vetements, camp.Label())
	if err != nil {
		return nil, "", err
	}
	return content, fmt.Sprintf("Liste de vêtements - %s.pdf", camp.Label()), nil
}
