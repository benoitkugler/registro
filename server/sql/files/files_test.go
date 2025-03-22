package files

import (
	"os"
	"testing"
	"time"

	tu "registro/utils/testutils"
)

func TestFilepath(t *testing.T) {
	tu.Assert(t, IdFile(4).filepath("root", false) == "root/file_4")
	tu.Assert(t, IdFile(4).filepath("root", true) == "root/file_4_min")
}

func TestUploadFile(t *testing.T) {
	fs := NewFileSystem(t.TempDir())
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	file, err := File{}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = UploadFile(fs, db, file.Id, nil, "")
	tu.AssertErr(t, err) // error for miniature

	f, err := os.ReadFile("test/img1.png")
	tu.AssertNoErr(t, err)
	meta, err := UploadFile(fs, db, file.Id, f, "img1.png")
	tu.AssertNoErr(t, err)

	tu.Assert(t, meta.Taille == len(f))
	tu.Assert(t, meta.Uploaded.YearDay() == time.Now().YearDay())
}
