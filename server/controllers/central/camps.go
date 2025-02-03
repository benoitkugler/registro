package central

import (
	"database/sql"
	"log"
	"time"

	cp "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type CampStats struct{}

type CampExt struct {
	Camp cp.Camp
	Taux ds.Taux
}

func (ct *Controller) CampsGet(c echo.Context) error {
	out, err := ct.loadCamps()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadCamps() ([]CampExt, error) {
	camps, err := cp.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	taux, err := ds.SelectTauxs(ct.db, camps.IdTauxs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	out := make([]CampExt, 0, len(camps))
	for _, camp := range camps {
		out = append(out, CampExt{camp, taux[camp.IdTaux]})
	}
	return out, nil
}

func (ct *Controller) CampsCreate(c echo.Context) error {
	out, err := ct.createCamp()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createCamp() (CampExt, error) {
	// TODO: better taux handling
	const defaultTaux = ds.IdTaux(1)
	camp, err := cp.Camp{
		IdTaux: defaultTaux, DateDebut: shared.NewDateFrom(time.Now()), Duree: 1,
		Places: 40,
	}.Insert(ct.db)
	if err != nil {
		return CampExt{}, utils.SQLError(err)
	}
	taux, err := ds.SelectTaux(ct.db, camp.IdTaux)
	if err != nil {
		return CampExt{}, utils.SQLError(err)
	}
	return CampExt{camp, taux}, nil
}

func (ct *Controller) CampsUpdate(c echo.Context) error {
	var args cp.Camp
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.updateCamp(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) updateCamp(args cp.Camp) (cp.Camp, error) {
	camp, err := cp.SelectCamp(ct.db, args.Id)
	if err != nil {
		return cp.Camp{}, utils.SQLError(err)
	}
	if err = args.Check(); err != nil {
		return cp.Camp{}, err
	}
	camp.Nom = args.Nom
	camp.DateDebut = args.DateDebut
	camp.Duree = args.Duree
	camp.Agrement = args.Agrement
	camp.Description = args.Description
	camp.Navette = args.Navette
	camp.Places = args.Places
	camp.Ouvert = args.Ouvert
	camp.Prix = args.Prix
	camp.OptionPrix = args.OptionPrix
	camp.OptionQuotientFamilial = args.OptionQuotientFamilial
	camp, err = camp.Update(ct.db)
	if err != nil {
		return cp.Camp{}, utils.SQLError(err)
	}

	return camp, nil
}

// CampsDelete supprime un camp SANS participants ni équipiers,
// renvoie une erreur sinon.
func (ct *Controller) CampsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cp.IdCamp](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteCamp(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteCamp(id cp.IdCamp) error {
	var toDelete []fs.IdFile
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		// intégrité sur les participants et équipiers

		links, err := fs.DeleteFileCampsByIdCamps(tx, id)
		if err != nil {
			return err
		}

		toDelete, err = fs.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}

		// cascade sur les DemandeCamps et Groupes
		_, err = cp.DeleteCampById(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// cleanup the document content
	go func() {
		err = ct.files.Delete(toDelete...)
		if err != nil {
			log.Println(err)
		}
	}()

	return nil
}

type CreateEquipierIn struct {
	IdPersonne pr.IdPersonne
	IdCamp     cp.IdCamp
}

func (ct *Controller) CampCreateEquipier(c echo.Context) error {
	var args CreateEquipierIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createEquipier(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createEquipier(args CreateEquipierIn) (out cp.Equipier, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		out, err = cp.Equipier{IdCamp: args.IdCamp, IdPersonne: args.IdPersonne}.Insert(tx)
		if err != nil {
			return err
		}
		demandes := ct.builtins.Defaut(out)
		err = fs.InsertManyDemandeEquipiers(tx, demandes...)
		if err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (ct *Controller) CampUpdateEquipier(c echo.Context) error {
	var args cp.Equipier
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateEquipier(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateEquipier(args cp.Equipier) error {
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}
