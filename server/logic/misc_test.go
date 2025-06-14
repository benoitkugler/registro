package logic

import (
	"database/sql"
	"testing"
	"time"

	cps "registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestController_searchSimilaires(t *testing.T) {
	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	err := utils.InTx(db.DB, func(tx *sql.Tx) error {
		for range [2000]int{} {
			_, err := pr.Personne{}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, err = SearchSimilaires(db, 1)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Since(ti) < 50*time.Millisecond)
}

func TestCheckPersonneReferences(t *testing.T) {
	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := dossiers.Dossier{IdResponsable: pe2.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	p1, err := cps.Participant{IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	p2, err := cps.Participant{IdCamp: camp2.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ref, err := CheckPersonneReferences(db, pe.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(ref.Participants) == 2)

	_, err = cps.DeleteParticipantById(db, p1.Id)
	tu.AssertNoErr(t, err)
	ref, err = CheckPersonneReferences(db, pe.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(ref.Participants) == 1)

	_, err = cps.DeleteParticipantById(db, p2.Id)
	tu.AssertNoErr(t, err)
	ref, err = CheckPersonneReferences(db, pe.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, ref.Empty())
}
