package central

import (
	"testing"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	"registro/sql/files"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestCRUD(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql",
	)
	defer db.Remove()

	pe, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	t.Run("camps", func(t *testing.T) {
		camps, err := ct.loadCamps()
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(camps) == 0)

		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)

		camps, err = ct.loadCamps()
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(camps) == 1)

		camp.Camp.Camp.Ouvert = true
		camp.Camp.Camp.Navette = cps.Navette{Actif: true, Commentaire: "7 € Aller retour"}
		updated, err := ct.updateCamp(camp.Camp.Camp)
		tu.AssertNoErr(t, err)
		tu.Assert(t, updated.Camp.Ouvert)

		err = ct.deleteCamp(camp.Camp.Camp.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("equipiers", func(t *testing.T) {
		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)

		_, err = ct.createEquipier(CreateEquipierIn{pe.Id, camp.Camp.Camp.Id})
		tu.AssertNoErr(t, err)
	})
}
