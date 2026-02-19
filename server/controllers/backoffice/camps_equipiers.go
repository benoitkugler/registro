package backoffice

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	fsAPI "registro/controllers/files"
	"registro/generators/sheets"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type CreateEquipierIn struct {
	IdPersonne pr.IdPersonne
	IdCamp     cps.IdCamp
	Roles      cps.Roles // to select default Demandes
}

func (ct *Controller) CampsCreateEquipier(c echo.Context) error {
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

func (ct *Controller) createEquipier(args CreateEquipierIn) (out cps.Equipier, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		out, err = cps.Equipier{IdCamp: args.IdCamp, IdPersonne: args.IdPersonne, Roles: args.Roles}.Insert(tx)
		if err != nil {
			return err
		}
		demandes := ct.builtins.Defaut(out.Id, out.Roles)
		err = fs.InsertManyDemandeEquipiers(tx, demandes...)
		if err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (ct *Controller) CampUpdateEquipier(c echo.Context) error {
	var args cps.Equipier
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateEquipier(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateEquipier(args cps.Equipier) error {
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) CampsDownloadEquipiers(c echo.Context) error {
	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	content, name, err := ct.exportListeEquipiers(idCamp)
	if err != nil {
		return err
	}
	mimeType := fsAPI.SetBlobHeader(c, content, name)
	return c.Blob(200, mimeType, content)
}

// export equpiers for the given camp
func (ct *Controller) exportListeEquipiers(idCamp cps.IdCamp) ([]byte, string, error) {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	equipiers, personnes, _, err := cps.LoadEquipiersByCamps(ct.db, idCamp)
	if err != nil {
		return nil, "", err
	}

	header := [...]string{"Nom", "Prénom", "Sexe", "Date de naissance", "Rôle", "Adresse", "Code postal", "Ville", "Pays", "Mail"}
	rows := make([][]sheets.Cell, 0, len(equipiers))
	for _, equipier := range equipiers {
		pe := personnes[equipier.IdPersonne]
		row := [len(header)]sheets.Cell{
			{Value: pe.FNom()},
			{Value: pe.FPrenom()},
			{Value: pe.Sexe.String()},
			{Value: pe.DateNaissance.String()},
			{Value: equipier.Roles.String()},
			{Value: pe.Adresse},
			{Value: pe.CodePostal},
			{Value: pe.Ville},
			{Value: string(pe.Pays)},
			{Value: pe.Mail},
		}
		rows = append(rows, row[:])
	}

	slices.SortFunc(rows, func(a, b []sheets.Cell) int { return strings.Compare(a[0].Value, b[0].Value) })

	content, err := sheets.CreateTable(header[:], rows)
	if err != nil {
		return nil, "", err
	}
	name := fmt.Sprintf("Equipiers %s.xlsx", camp.Label())
	return content, name, nil
}
