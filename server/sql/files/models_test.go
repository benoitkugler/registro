package files

import (
	"testing"

	"registro/sql/camps"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	files, err := SelectAllFiles(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 0)

	pers, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{}.Insert(db)
	tu.AssertNoErr(t, err)
	file1, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	file2, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	file3, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)

	err = InsertFileCamp(db, FileCamp{IdCamp: camp.Id, IdFile: file1.Id, IsLettre: true})
	tu.AssertNoErr(t, err)
	err = InsertFileCamp(db, FileCamp{IdCamp: camp.Id, IdFile: file2.Id, IsLettre: true})
	tu.Assert(t, err != nil) // lettre is unique
	err = InsertFileCamp(db, FileCamp{IdCamp: camp.Id, IdFile: file2.Id, IsLettre: false})
	tu.AssertNoErr(t, err)
	err = InsertFileCamp(db, FileCamp{IdCamp: camp.Id, IdFile: file3.Id, IsLettre: false})
	tu.AssertNoErr(t, err)

	// demandes
	_, err = Demande{MaxDocs: 1, Categorie: Vaccin}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = Demande{MaxDocs: 1, Categorie: Vaccin}.Insert(db)
	tu.Assert(t, err != nil) // unique
	_, err = Demande{MaxDocs: 1, Categorie: 0}.Insert(db)
	tu.Assert(t, err != nil) // missing directeur
	_, err = Demande{MaxDocs: 1, Categorie: 0, IdDirecteur: OptIdPersonne{Id: pers.Id, Valid: true}}.Insert(db)
	tu.AssertNoErr(t, err)
}
