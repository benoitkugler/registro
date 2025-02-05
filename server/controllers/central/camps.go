package central

import (
	"database/sql"
	"errors"
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

func (ct *Controller) CampsGetTaux(c echo.Context) error {
	out, err := ct.loadTaux()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadTaux() (ds.Tauxs, error) {
	return ds.SelectAllTauxs(ct.db)
}

type CampHeader struct {
	Camp  cp.Camp
	Taux  ds.Taux
	Stats cp.StatistiquesInscrits
}

func (ct *Controller) CampsGet(c echo.Context) error {
	out, err := ct.loadCamps()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadCamps() ([]CampHeader, error) {
	camps, err := cp.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	taux, err := ds.SelectTauxs(ct.db, camps.IdTauxs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	participants, err := cp.SelectParticipantsByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	byCamp := participants.ByIdCamp()
	personnes, err := pr.SelectPersonnes(ct.db, participants.IdPersonnes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	out := make([]CampHeader, 0, len(camps))
	for _, camp := range camps {
		loader := cp.CampLoader{Camp: &camp, Participants: byCamp[camp.Id], Personnes: personnes}
		out = append(out, CampHeader{camp, taux[camp.IdTaux], loader.Stats()})
	}
	return out, nil
}

func ensureTaux(tx *sql.Tx, taux ds.Taux) (ds.Taux, error) {
	if taux.Id <= 0 { // create a new Taux
		return taux.Insert(tx)
	} // else, simply use the Id
	return taux, nil
}

type CampsCreateManyIn struct {
	// If Taux has an [Id] <= 0, it is created first,
	// Otherwise, only its [Id] is used.
	Taux ds.Taux
	// Count is the number of Camp to create
	Count int
}

func (ct *Controller) CampsCreateMany(c echo.Context) error {
	var args CampsCreateManyIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createManyCamp(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createManyCamp(args CampsCreateManyIn) (out []CampHeader, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		args.Taux, err = ensureTaux(tx, args.Taux)
		if err != nil {
			return err
		}
		for i := 0; i < args.Count; i++ {
			camp, err := defaultCamp(args.Taux.Id).Insert(tx)
			if err != nil {
				return err
			}
			out = append(out, CampHeader{camp, args.Taux, cp.StatistiquesInscrits{}})
		}
		return nil
	})
	return out, err
}

func (ct *Controller) CampsCreate(c echo.Context) error {
	out, err := ct.createCamp()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func defaultCamp(idTaux ds.IdTaux) cp.Camp {
	return cp.Camp{
		IdTaux:    idTaux,
		Nom:       "Nouveau séjour",
		DateDebut: shared.NewDateFrom(time.Now()), Duree: 1,
		Places: 40, AgeMin: 6, AgeMax: 12,
	}
}

func (ct *Controller) createCamp() (CampHeader, error) {
	const defaultTaux = ds.IdTaux(1) // always present in the DB, by construction
	camp, err := defaultCamp(defaultTaux).Insert(ct.db)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}
	taux, err := ds.SelectTaux(ct.db, camp.IdTaux)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}
	return CampHeader{camp, taux, cp.StatistiquesInscrits{}}, nil
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
	camp.Lieu = args.Lieu
	camp.Agrement = args.Agrement
	camp.Description = args.Description
	camp.Navette = args.Navette
	camp.Places = args.Places
	camp.AgeMin = args.AgeMin
	camp.AgeMax = args.AgeMax
	camp.Ouvert = args.Ouvert
	camp.Prix = args.Prix
	camp.OptionPrix = args.OptionPrix
	camp.OptionQuotientFamilial = args.OptionQuotientFamilial
	camp.Password = args.Password
	camp, err = camp.Update(ct.db)
	if err != nil {
		return cp.Camp{}, utils.SQLError(err)
	}

	return camp, nil
}

type CampsSetTauxIn struct {
	IdCamp cp.IdCamp
	// If Taux has an [Id] <= 0, it is created first,
	// Otherwise, only its [Id] is used.
	Taux ds.Taux
}

// CampsSetTaux permet à l'utilisateur de changer le Taux
// utilisé par un séjour donné, à condition qu'il n'y ait
// pas encore de participants.
func (ct *Controller) CampsSetTaux(c echo.Context) error {
	var args CampsSetTauxIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.setTaux(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) setTaux(args CampsSetTauxIn) (out CampHeader, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		participants, err := cp.SelectParticipantsByIdCamps(tx, args.IdCamp)
		if err != nil {
			return err
		}
		if len(participants) > 0 {
			return errors.New("des participants sont déjà déclarés")
		}

		args.Taux, err = ensureTaux(tx, args.Taux)
		if err != nil {
			return err
		}
		camp, err := cp.SelectCamp(tx, args.IdCamp)
		if err != nil {
			return err
		}
		camp.IdTaux = args.Taux.Id
		camp, err = camp.Update(tx)
		if err != nil {
			return err
		}
		out = CampHeader{camp, args.Taux, cp.StatistiquesInscrits{}} // no participants
		return nil
	})
	return out, err
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
		// intégrité sur les participants
		// cascade sur les équipiers

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
