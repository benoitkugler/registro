package files

import (
	"testing"

	"registro/sql/camps"
	"registro/sql/dossiers"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	files, err := SelectAllFiles(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 0)

	pers, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	taux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	file1, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	file2, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	file3, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)

	err = FileCamp{IdCamp: camp.Id, IdFile: file1.Id, IsLettre: true}.Insert(db)
	tu.AssertNoErr(t, err)
	err = FileCamp{IdCamp: camp.Id, IdFile: file2.Id, IsLettre: true}.Insert(db)
	tu.AssertErr(t, err) // lettre is unique
	err = FileCamp{IdCamp: camp.Id, IdFile: file2.Id, IsLettre: false}.Insert(db)
	tu.AssertNoErr(t, err)
	err = FileCamp{IdCamp: camp.Id, IdFile: file3.Id, IsLettre: false}.Insert(db)
	tu.AssertNoErr(t, err)

	// demandes
	_, err = Demande{MaxDocs: 1, Categorie: Vaccins}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = Demande{MaxDocs: 1, Categorie: Vaccins}.Insert(db)
	tu.AssertErr(t, err) // unique
	_, err = Demande{MaxDocs: 1, Categorie: 0}.Insert(db)
	tu.AssertNoErr(t, err) // shared constraints
	_, err = Demande{MaxDocs: 1, Categorie: 0, IdDirecteur: OptIdPersonne{Id: pers.Id, Valid: true}}.Insert(db)
	tu.AssertNoErr(t, err)
}
