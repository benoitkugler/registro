package logic

import (
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestIdentifieProfil(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	fi, err := files.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	demande, err := files.Demande{MaxDocs: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = files.FilePersonne{IdFile: fi.Id, IdPersonne: pe1.Id, IdDemande: demande.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = IdentifiePersonne(db.DB, IdentTarget{
		IdTemporaire: pe1.Id,
		Rattache:     true,
		RattacheTo:   pe2.Id,
	})
	tu.AssertNoErr(t, err)

	links, err := files.SelectFilePersonnesByIdPersonnes(db, pe2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 1)
}
