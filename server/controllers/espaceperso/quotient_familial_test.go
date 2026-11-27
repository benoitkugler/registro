package espaceperso

import (
	"testing"

	"registro/config"
	"registro/crypto"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func Test_uploadQuotientFamilial(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier2, err := ds.Dossier{IdTaux: 1, IdResponsable: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(t.TempDir()), config.Immich{}, nil, nil)

	file, err := ct.uploadQuotientFamilial(dossier.Id, tu.PngData, "attestation.png")
	tu.AssertNoErr(t, err)
	tu.Assert(t, file.Key != "")

	link, found, err := fs.SelectFileQuotientFamilialByIdDossier(db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && link.IdFile == file.Id)
	firstFile := link.IdFile

	// une attestation validée ne peut être ni remplacée ni supprimée
	stored, err := fs.SelectFile(db, link.IdFile)
	tu.AssertNoErr(t, err)
	stored.ReviewStatus = pr.FileReviewValidated
	_, err = stored.Update(db)
	tu.AssertNoErr(t, err)

	_, err = ct.uploadQuotientFamilial(dossier.Id, tu.PngData, "attestation2.png")
	tu.AssertErr(t, err)
	err = ct.deleteQuotientFamilial(dossier.Id)
	tu.AssertErr(t, err)
	link, found, err = fs.SelectFileQuotientFamilialByIdDossier(db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && link.IdFile == firstFile)
	stored, err = fs.SelectFile(db, link.IdFile)
	tu.AssertNoErr(t, err)
	tu.Assert(t, stored.ReviewStatus == pr.FileReviewValidated)

	// après refus par le backoffice, un nouvel envoi est à nouveau possible
	stored.ReviewStatus = pr.FileReviewRejected
	_, err = stored.Update(db)
	tu.AssertNoErr(t, err)

	file2, err := ct.uploadQuotientFamilial(dossier.Id, tu.PngData, "attestation2.png")
	tu.AssertNoErr(t, err)
	link, found, err = fs.SelectFileQuotientFamilialByIdDossier(db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && link.IdFile == file2.Id && link.IdFile != firstFile)
	links, err := fs.SelectFileQuotientFamilialsByIdDossiers(db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 1)
	stored, err = fs.SelectFile(db, link.IdFile)
	tu.AssertNoErr(t, err)
	tu.Assert(t, stored.ReviewStatus == pr.FileReviewPending)

	err = ct.deleteQuotientFamilial(dossier.Id)
	tu.AssertNoErr(t, err)
	_, found, err = fs.SelectFileQuotientFamilialByIdDossier(db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, !found)

	// supprimer un dossier sans attestation est un no-op
	err = ct.deleteQuotientFamilial(dossier2.Id)
	tu.AssertNoErr(t, err)
}
