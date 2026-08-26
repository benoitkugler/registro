package backoffice

import (
	"database/sql"
	"errors"

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

// PersonnesLoadFichesanitaire renvoie la fiche sanitaire,
// avec en particulier les droits d'accès.
func (ct *Controller) PersonnesLoadFichesanitaire(c echo.Context) error {
	id, err := utils.QueryParamInt[pr.IdPersonne](c, "id")
	if err != nil {
		return err
	}
	out, _, err := pr.SelectFichesanitaireByIdPersonne(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.JSON(200, out)
}

type UpdateFichesanitaireAccessIn struct {
	Id    pr.IdPersonne
	Mails pr.Mails
}

func (ct *Controller) PersonnesUpdateFichesanitaireAccess(c echo.Context) error {
	var args UpdateFichesanitaireAccessIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateFichesanitaireAccess(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateFichesanitaireAccess(args UpdateFichesanitaireAccessIn) error {
	if len(args.Mails) == 0 {
		return errors.New("internal error: at least one Mail is required")
	}

	fiche, found, err := pr.SelectFichesanitaireByIdPersonne(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if !found {
		return errors.New("internal error: missing Fichesanitaire")
	}

	fiche.Owners = args.Mails

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		err = fiche.Delete(tx)
		if err != nil {
			return err
		}
		err = fiche.Insert(tx)
		return err
	})
}
