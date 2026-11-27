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
	tu.AssertErr(t, err) // unique
	_, err = Demande{MaxDocs: 1, Categorie: NoBuiltin}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = Demande{MaxDocs: 1, Categorie: NoBuiltin}.Insert(db)
	tu.AssertNoErr(t, err) // shared constraints
	_, err = Demande{MaxDocs: 1, Categorie: 0, IdDirecteur: pers.Id.Opt()}.Insert(db)
	tu.AssertNoErr(t, err)
}

func TestFileAide(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	pers, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	taux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := dossiers.Dossier{IdResponsable: pers.Id, IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	part, err := camps.Participant{IdCamp: camp.Id, IdPersonne: pers.Id, IdDossier: dossier.Id, IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	structure, err := camps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)
	aide, err := camps.Aide{IdParticipant: part.Id, IdStructureaide: structure.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	file, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	// un justificatif peut exister avant l'aide
	err = FileAide{IdFile: file.Id, IdParticipant: part.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	links, err := SelectFileAidesByIdParticipants(db, part.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 1 && !links[0].IdAide.Valid)

	link, found, err := SelectFileAideByIdFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && !link.IdAide.Valid)

	// rattachement en deux étapes avec les méthodes générées
	_, err = DeleteFileAidesByIdFiles(db, link.IdFile)
	tu.AssertNoErr(t, err)
	err = FileAide{
		IdFile:        link.IdFile,
		IdParticipant: link.IdParticipant,
		IdAide:        aide.Id.Opt(),
	}.Insert(db)
	tu.AssertNoErr(t, err)
	link, found, err = SelectFileAideByIdFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && link.IdAide.Is(aide.Id))

	// une seule pièce peut être rattachée à une aide donnée
	file2, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = FileAide{IdFile: file2.Id, IdParticipant: part.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	link2, found, err := SelectFileAideByIdFile(db, file2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && !link2.IdAide.Valid)
	_, err = DeleteFileAidesByIdFiles(db, link2.IdFile)
	tu.AssertNoErr(t, err)
	err = FileAide{
		IdFile:        link2.IdFile,
		IdParticipant: link2.IdParticipant,
		IdAide:        aide.Id.Opt(),
	}.Insert(db)
	tu.AssertErr(t, err)
}
