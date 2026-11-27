package backoffice

import (
	"os"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestController_quotientFamilial(t *testing.T) {
	content, err := os.ReadFile("../../caf/test/attestation_quotient.pdf")
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdResponsable: pe.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	part, err := cps.Participant{IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, key: crypto.Encrypter{}, files: fs.NewFileSystem(t.TempDir()), asso: config.Asso{}}

	file, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileQuotientFamilial{IdFile: file.Id, IdDossier: dossier.Id}).Insert(db)
	tu.AssertNoErr(t, err)
	_, err = fs.UploadFile(ct.files, db, file, content, "attestation.pdf")
	tu.AssertNoErr(t, err)

	// la détection est faite au chargement du dossier
	details, err := ct.loadDossier("localhost", dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, details.Dossier.HasQuotientFamilialFile)
	tu.Assert(t, details.QuotientFamilialDetection.Status == QFDetectionSuccess)
	tu.Assert(t, details.QuotientFamilialDetection.QuotientFamilial > 0)
	tu.Assert(t, details.QuotientFamilialDetection.Annee >= 2000)

	// la validation applique la valeur (éventuellement corrigée) aux participants
	err = ct.validateQuotientFamilial(QuotientFamilialValidateIn{
		IdDossier: dossier.Id, QuotientFamilial: 1234, Accepted: true,
	})
	tu.AssertNoErr(t, err)
	updated, err := cps.SelectParticipant(db, part.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, updated.QuotientFamilial == 1234)
	fileDB, err := fs.SelectFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, fileDB.ReviewStatus == pr.FileReviewValidated)

	// un refus ne modifie pas le quotient familial
	err = ct.validateQuotientFamilial(QuotientFamilialValidateIn{
		IdDossier: dossier.Id, QuotientFamilial: 500, Accepted: false,
	})
	tu.AssertNoErr(t, err)
	updated, err = cps.SelectParticipant(db, part.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, updated.QuotientFamilial == 1234)
	fileDB, err = fs.SelectFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, fileDB.ReviewStatus == pr.FileReviewRejected)

	// un PDF invalide est signalé sans empêcher le chargement du dossier
	_, err = fs.UploadFile(ct.files, db, file, tu.PngData, "scan.png")
	tu.AssertNoErr(t, err)
	details, err = ct.loadDossier("localhost", dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, details.QuotientFamilialDetection.Status == QFDetectionInvalidPDF)
	tu.Assert(t, details.QuotientFamilialDetection.Message != "")
}
