package compta

import (
	"testing"
	"time"

	"registro/sql/camps"
	"registro/sql/dossiers"
	tu "registro/utils/testutils"
)

func TestSQLRandom(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	defautTaux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: defautTaux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	compta, err := ComptaCamp{IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ComptaCamp{IdCamp: camp.Id}.Insert(db)
	tu.AssertErr(t, err)

	for range 10 {
		be := randComptaRecord()
		be.SpentBy = OptIdEquipier{}
		be.IdCamp = compta.Id
		be, err = be.Insert(db)
		tu.AssertNoErr(t, err)

		be.SpentBy = camps.IdEquipier(10).Opt() // does not exist
		_, err = be.Update(db)
		tu.AssertErr(t, err)

		rec := randComptaSubrecord()
		rec.IdRecord = be.Id
		rec.IdFile = OptIdComptaFile{}
		_, err = rec.Insert(db)
		tu.AssertNoErr(t, err)
	}
}

func TestSQLCompta(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	defautTaux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)

	cs, err := SelectAllComptaCamps(db.DB)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(cs) == 0)

	camp, err := camps.Camp{IdTaux: defautTaux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	cc, err := ComptaCamp{IdCamp: camp.Id, Account: Clairiere, State: Archive}.Insert(db)
	tu.AssertNoErr(t, err)

	cs, err = SelectAllComptaCamps(db.DB)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(cs) == 1)

	rec, err := ComptaRecord{
		IdCamp: cc.Id,
		Reviews: Reviews{
			{true, time.Now(), "A first message"},
			{false, time.Now(), "A second answer"},
		},
	}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ComptaSubrecord{
		IdRecord: rec.Id,
		Currency: dossiers.Euros,
		Montants: Montants{
			{1000, Alimentation},
			{80, Bank},
		},
	}.Insert(db)
	tu.AssertNoErr(t, err)

	recs, err := SelectComptaRecordsByIdCamps(db, cc.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(recs) == 1)

	_, err = camps.DeleteCampById(db, camp.Id)
	tu.AssertNoErr(t, err)

	cs, err = SelectAllComptaCamps(db.DB)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(cs) == 0)
}

func TestBudget(t *testing.T) {
	db := tu.NewTestDB(t, "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	defautTaux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)

	for range 10 {
		camp, err := camps.Camp{IdTaux: defautTaux.Id}.Insert(db)
		tu.AssertNoErr(t, err)

		cc, err := ComptaCamp{IdCamp: camp.Id}.Insert(db)
		tu.AssertNoErr(t, err)

		budget := randComptaBudget()
		budget.IdCompta = cc.Id
		err = budget.Insert(db)
		tu.AssertNoErr(t, err)

		config := randComptaConfig()
		_, err = config.Insert(db)
		tu.AssertNoErr(t, err)
	}
}
