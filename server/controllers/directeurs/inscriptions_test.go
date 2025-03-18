package directeurs

import (
	"reflect"
	"testing"
	"time"

	"registro/controllers/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestValideInscription(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp2.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	err = ct.valideInscription(dossier1.Id, camp1.Id)
	tu.AssertNoErr(t, err)
	data, err := logic.LoadDossier(db, dossier1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, !data.Dossier.IsValidated)

	insc, err := ct.getInscriptions(camp1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(insc) == 1 && reflect.DeepEqual(insc[0].ValidatedBy, []cps.IdCamp{camp1.Id}))

	err = ct.valideInscription(dossier1.Id, camp2.Id)
	tu.AssertNoErr(t, err)
	data, err = logic.LoadDossier(db, dossier1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.Dossier.IsValidated)
}
