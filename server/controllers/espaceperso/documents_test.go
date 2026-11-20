package espaceperso

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

func TestDocuments(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.NewDateFrom(tu.DateFor(25))}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.NewDateFrom(tu.DateFor(8))}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe4, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{
		IdTaux: 1, DateDebut: shared.NewDateFrom(time.Now()),
		DocumentsReady:  true,
		DocumentsToShow: cps.DocumentsToShow{ListeVetements: true, CharteParticipant: true},
	}.Insert(db)
	tu.AssertNoErr(t, err)

	customD, err := fs.Demande{MaxDocs: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	err = fs.DemandeCamp{IdDemande: customD.Id, IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	form, err := cps.Form{IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	p1, err := cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe1.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	p2, err := cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe2.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe3.Id, IdDossier: dossier.Id, Statut: cps.AStatuer}.Insert(db)
	tu.AssertNoErr(t, err)

	err = cps.ParticipantForm{IdForm: form.Id, IdCamp: camp.Id, IdParticipant: p1.Id, Reponses: cps.FormReponses{cps.ChampReponseTexte("")}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{AskedDocs: config.AskedDocs{CarteId: true}}, fs.NewFileSystem(os.TempDir()), config.Immich{})

	docs, err := ct.markAndloadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(docs.FilesToRead) == 1)
	tu.Assert(t, len(docs.FilesToRead[0].Files) == 0 && len(docs.FilesToRead[0].Generated) == 1)
	tu.Assert(t, len(docs.FilesToUpload) == 2)
	tu.Assert(t, len(docs.FilesToUpload[0].Demandes[0].Uploaded) == 0)
	tu.Assert(t, len(docs.Chartes) == 1)
	tu.Assert(t, len(docs.Fiches) == 1)
	tu.Assert(t, len(docs.Forms) == 2)
	tu.Assert(t, docs.NewCount == 0+5+1+1+1)

	for _, personne := range docs.FilesToUpload {
		hasCarteId, hasVaccins := false, false
		for _, demande := range personne.Demandes {
			hasCarteId = hasCarteId || demande.Demande.Categorie == fs.CarteId
			hasVaccins = hasVaccins || demande.Demande.Categorie == fs.Vaccins
		}
		tu.Assert(t, hasCarteId)
		if personne.IdPersonne == pe1.Id {
			tu.Assert(t, !hasVaccins)
		}
	}

	ct.asso.AskedDocs.CarteId = false
	docsWithoutCarteId, err := ct.markAndloadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	for _, personne := range docsWithoutCarteId.FilesToUpload {
		for _, demande := range personne.Demandes {
			tu.Assert(t, demande.Demande.Categorie != fs.CarteId)
		}
	}
	ct.asso.AskedDocs.CarteId = true

	_, err = ct.uploadDocument(dossier.Id, customD.Id, pe4.Id, tu.PngData, "test.png")
	tu.AssertErr(t, err)

	_, err = ct.uploadDocument(dossier.Id, customD.Id, pe1.Id, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)

	err = ct.accepteCharte(dossier.Id, pe1.Id)
	tu.AssertNoErr(t, err)

	docs, err = ct.markAndloadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(docs.FilesToUpload[0].Demandes) == 2)
	tu.Assert(t, len(docs.FilesToUpload[0].Demandes[1].Uploaded) == 1)
	tu.Assert(t, docs.NewCount == 3+1+1+1)

	err = cps.ParticipantForm{IdForm: form.Id, IdCamp: camp.Id, IdParticipant: p2.Id, Reponses: cps.FormReponses{cps.ChampReponseTexte("")}}.Insert(db)
	tu.AssertNoErr(t, err)
	docs, err = ct.markAndloadDocuments(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, docs.NewCount == 3+1+1+0)
}
