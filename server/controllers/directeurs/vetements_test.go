package directeurs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	"registro/sql/files"
	tu "registro/utils/testutils"
)

func TestVetements(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	err = pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: files.NewFileSystem(os.TempDir())}

	_, err = ct.getVetements(camp1.Id)
	tu.AssertNoErr(t, err)

	liste := make([]cps.Vetement, 30)
	liste[0] = cps.Vetement{Quantite: 2, Description: "e", Important: true}
	liste[1] = cps.Vetement{Quantite: 2, Description: "e", Important: false}
	err = ct.updateVetements(camp1.Id, cps.ListeVetements{
		Vetements:  liste,
		Complement: "<b>smdsd</b> <a></a>",
	})
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, _, err = ct.renderListeVetements(camp1.Id)
	tu.AssertNoErr(t, err)
	fmt.Println("rendered in", time.Since(ti))
}
