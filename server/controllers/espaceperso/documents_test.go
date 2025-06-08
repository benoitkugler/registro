package espaceperso

import (
	"os"
	"testing"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestDocuments(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe4, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1, DocumentsToShow: cps.DocumentsToShow{ListeVetements: true}}.Insert(db)
	tu.AssertNoErr(t, err)

	d1, err := fs.Demande{MaxDocs: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	err = fs.DemandeCamp{IdDemande: d1.Id, IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe1.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe2.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe3.Id, IdDossier: dossier.Id, Statut: cps.AStatuer}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(os.TempDir()), config.Joomeo{})

	docs, err := ct.loadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(docs.FilesToRead) == 1)
	tu.Assert(t, len(docs.FilesToUpload) == 2)
	tu.Assert(t, len(docs.FilesToUpload[0].Demandes[0].Uploaded) == 0)
	tu.Assert(t, docs.ToFillCount == 2)

	_, err = ct.uploadDocument(dossier.Id, d1.Id, pe4.Id, tu.PngData, "test.png")
	tu.AssertErr(t, err)

	_, err = ct.uploadDocument(dossier.Id, d1.Id, pe1.Id, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)

	docs, err = ct.loadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(docs.FilesToUpload[0].Demandes[0].Uploaded) == 1)
	tu.Assert(t, docs.ToFillCount == 1)
}
