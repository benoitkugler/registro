package directeurs

import (
	"os"
	"testing"

	"registro/config"
	filesAPI "registro/controllers/files"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func createFileFor(db ds.DB, idPersonne pr.IdPersonne, idDemande fs.IdDemande) error {
	file, err := fs.File{}.Insert(db)
	if err != nil {
		return err
	}
	err = fs.FilePersonne{IdFile: file.Id, IdPersonne: idPersonne, IdDemande: idDemande}.Insert(db)
	if err != nil {
		return err
	}
	return nil
}

func TestDocuments(t *testing.T) {
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

	pe3, err := pr.Personne{}.Insert(db) // not in the camp
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdPersonne: pe1.Id, IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdPersonne: pe2.Id, IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, "", "", files.NewFileSystem(os.TempDir()), config.SMTP{}, config.Asso{}, config.Joomeo{})
	tu.AssertNoErr(t, err)

	t.Run("files to download", func(t *testing.T) {
		file1, err := ct.uploadToDownload(camp.Id, tu.PngData, "test.png")
		tu.AssertNoErr(t, err)

		_, err = ct.uploadToDownload(camp.Id, tu.PngData, "test2.png")
		tu.AssertNoErr(t, err)

		err = ct.updateToShow(camp.Id, cps.DocumentsToShow{LettreDirecteur: true, ListeVetements: false, ListeParticipants: true})
		tu.AssertNoErr(t, err)

		docs, err := ct.getDocuments(camp.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(docs.FilesToDownload) == 2)
		tu.Assert(t, docs.ToShow.LettreDirecteur)

		err = filesAPI.Delete(ct.db, ct.key, ct.files, file1.Key)
		tu.AssertNoErr(t, err)

		docs, err = ct.getDocuments(camp.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(docs.FilesToDownload) == 1)
	})

	t.Run("files to upload", func(t *testing.T) {
		demande, err := ct.createDemande(camp.Id)
		tu.AssertErr(t, err) // no directeur

		_, err = ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: true,
			Roles:          cps.Roles{cps.Direction},
		}, camp.Id)
		tu.AssertNoErr(t, err)

		demande, err = ct.createDemande(camp.Id)
		tu.AssertNoErr(t, err)

		demande.Demande.JoursValide = 30
		err = ct.updateDemande(camp.Id, demande.Demande)
		tu.AssertNoErr(t, err)

		_, err = ct.uploadDemandeFile(camp.Id, demande.Demande.Id, tu.PngData, "test.png")
		tu.AssertNoErr(t, err)

		_, err = ct.applyDemande(camp.Id, demande.Demande.Id)
		tu.AssertNoErr(t, err)

		err = ct.unapplyDemande(camp.Id, demande.Demande.Id)
		tu.AssertNoErr(t, err)

		_, err = ct.applyDemande(camp.Id, demande.Demande.Id)
		tu.AssertNoErr(t, err)

		err = ct.deleteDemande(camp.Id, demande.Demande.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("retrieve files", func(t *testing.T) {
		demande1, err := ct.createDemande(camp.Id)
		tu.AssertNoErr(t, err)
		_, err = ct.applyDemande(camp.Id, demande1.Demande.Id)
		tu.AssertNoErr(t, err)

		demande2, err := ct.createDemande(camp.Id)
		tu.AssertNoErr(t, err)
		_, err = ct.applyDemande(camp.Id, demande2.Demande.Id)
		tu.AssertNoErr(t, err)

		err = createFileFor(ct.db, pe1.Id, demande1.Demande.Id)
		tu.AssertNoErr(t, err)
		err = createFileFor(ct.db, pe2.Id, demande1.Demande.Id)
		tu.AssertNoErr(t, err)
		err = createFileFor(ct.db, pe3.Id, demande1.Demande.Id)
		tu.AssertNoErr(t, err)

		err = createFileFor(ct.db, pe1.Id, demande2.Demande.Id)
		tu.AssertNoErr(t, err)
		err = createFileFor(ct.db, pe3.Id, demande2.Demande.Id)
		tu.AssertNoErr(t, err)

		out, err := ct.getUploaded(camp.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(out.Personnes) == 2)
		tu.Assert(t, len(out.DemandesDocuments) == 2)
		tu.Assert(t, len(out.DemandesDocuments[0].UploadedBy) == 2)
		tu.Assert(t, len(out.DemandesDocuments[1].UploadedBy) == 1)

		files, _, err := ct.selectDocumentsForDemande(camp.Id, demande1.Demande.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(files) == 2)
	})
}
