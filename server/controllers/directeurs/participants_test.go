package directeurs

import (
	"os"
	"testing"

	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestFichesSanitaires(t *testing.T) {
	err := pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	err = pr.Fichesanitaire{IdPersonne: pe1.Id, TraitementMedical: true}.Insert(db)
	tu.AssertNoErr(t, err)
	err = pr.Fichesanitaire{IdPersonne: pe2.Id, TraitementMedical: true}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	pa1, err := cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe3.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	fiches, err := ct.loadFichesSanitaires(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(fiches) == 3)

	_, _, err = ct.downloadFicheSanitaire(camp.Id, pa1.Id)
	tu.AssertNoErr(t, err)

	_, _, err = ct.downloadFichesSanitaires(camp.Id)
	tu.AssertNoErr(t, err)
}
