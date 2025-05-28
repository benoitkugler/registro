package directeurs

import (
	"fmt"
	"os"
	"testing"

	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	"registro/sql/files"
	tu "registro/utils/testutils"
)

func TestLettres(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	camp2, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: files.NewFileSystem(os.TempDir())}

	url1, err := ct.uploadLettreImage("localhost", tu.PngData, "test.png")
	tu.AssertNoErr(t, err)

	url2, err := ct.uploadLettreImage("localhost", tu.PngData, "test.png")
	tu.AssertNoErr(t, err)

	// both images
	err = cps.Lettredirecteur{IdCamp: camp1.Id, Html: fmt.Sprintf(`
	<body>
	<img src="%s" />
	<img src="%s" />
	</body>
	`, url1, url2)}.Insert(db)
	tu.AssertNoErr(t, err)

	// only 1
	err = cps.Lettredirecteur{IdCamp: camp2.Id, Html: fmt.Sprintf(`
	<body>
	<img src="%s" />
	</body>
	`, url1)}.Insert(db)
	tu.AssertNoErr(t, err)

	// nothing to remove
	err = ct.garbageCollectImages()
	tu.AssertNoErr(t, err)
	images, err := cps.SelectAllLettreImages(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(images) == 2)

	// remove lettre 1 : image 2 is unused
	_, err = cps.DeleteLettredirecteursByIdCamps(db, camp1.Id)
	tu.AssertNoErr(t, err)
	err = ct.garbageCollectImages()
	tu.AssertNoErr(t, err)
	images, err = cps.SelectAllLettreImages(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(images) == 1)

	err = pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)
	_, err = ct.updateLettreDirecteur(camp1.Id, cps.Lettredirecteur{
		UseCoordCentre: true,
		Html:           "Test2",
	})
	tu.AssertNoErr(t, err)

	_, err = ct.updateLettreDirecteur(camp1.Id, cps.Lettredirecteur{
		UseCoordCentre: false,
		Html:           "Test",
	})
	tu.AssertErr(t, err) // pas de directeur

	out, err := ct.getLettre(camp1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.Lettre.Html == "Test2")
}
