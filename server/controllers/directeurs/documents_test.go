package directeurs

import (
	"os"
	"testing"

	"registro/config"
	filesAPI "registro/controllers/files"
	cps "registro/sql/camps"
	"registro/sql/files"
	tu "registro/utils/testutils"
)

func TestDocuments(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, "", "", files.NewFileSystem(os.TempDir()), config.SMTP{}, config.Asso{}, config.Joomeo{})
	tu.AssertNoErr(t, err)

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
}
