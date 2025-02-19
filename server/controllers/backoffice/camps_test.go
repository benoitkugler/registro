package backoffice

import (
	"testing"

	"registro/config"
	"registro/crypto"
	cp "registro/sql/camps"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	"registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestCRUD(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	t.Run("camps", func(t *testing.T) {
		camps, err := ct.getCamps()
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(camps) == 0)

		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)

		camps, err = ct.getCamps()
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

func Test_lastTaux(t *testing.T) {
	tests := []struct {
		camps cp.Camps
		want  ds.IdTaux
	}{
		{nil, 1},
		{cps.Camps{2: {IdTaux: 2, DateDebut: shared.NewDate(2000, 1, 1)}}, 2},
		{cps.Camps{
			2: {IdTaux: 2, DateDebut: shared.NewDate(2000, 1, 1)},
			3: {IdTaux: 3, DateDebut: shared.NewDate(2001, 1, 1)},
		}, 3},
	}
	for _, tt := range tests {
		tu.Assert(t, lastTaux(tt.camps) == tt.want)
	}
}
