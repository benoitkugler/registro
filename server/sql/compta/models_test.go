package compta

import (
	"testing"

	"registro/sql/camps"
	"registro/sql/dossiers"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	defautTaux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: defautTaux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	compta, err := ComptaCamp{IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	be := randComptaBankentry()
	be.SpentBy = OptIdEquipier{}
	be.IdCamp = compta.Id
	be, err = be.Insert(db)
	tu.AssertNoErr(t, err)

	be.SpentBy = camps.IdEquipier(10).Opt() // does not exist
	_, err = be.Update(db)
	tu.AssertErr(t, err)

	rec := randComptaRecord()
	rec.IdBankentry = be.Id
	_, err = rec.Insert(db)
	tu.AssertNoErr(t, err)

	ca := randComptaCampACVE()
	ca.IdComptaCamp = compta.Id
	err = ca.Insert(db)
	tu.AssertNoErr(t, err)
}
