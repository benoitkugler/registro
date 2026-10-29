package compta

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"registro/config"
	ca "registro/sql/compta"
	tu "registro/utils/testutils"
	"testing"
)

func devEnv(t *testing.T) (googleDriveAPI, fileSystemAPI, config.SMTP) {
	tu.LoadEnv(t, "../../env.sh")
	st, err := config.NewCompta()
	tu.AssertNoErr(t, err)
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return googleDriveAPI{st.GoogleDriveKeyFile, st.GoogleDriveFolderId}, fileSystemAPI{st.FilesDir}, smtp
}

func TestGoogleDriveAuth(t *testing.T) {
	gd, _, _ := devEnv(t)
	tu.AssertNoErr(t, gd.check())
}

func TestDriveAPI(t *testing.T) {
	// t.Skip()
	api, _, _ := devEnv(t)

	i1, i2, err := api.createFile("test.png", tu.PngData)
	tu.AssertNoErr(t, err)

	b, err := api.download(i1)
	tu.AssertNoErr(t, err)
	tu.Assert(t, bytes.Equal(tu.PngData, b))

	b, err = api.download(i2)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(b) > 100)

	var content bytes.Buffer
	archive := zip.NewWriter(&content)
	err = api.downloadToZip([]idAndName{{1, "premier"}, {2, "deuxième"}}, ca.ComptaFiles{1: {DriveID: i1}, 2: {DriveID: i2}}, archive)
	tu.AssertNoErr(t, err)
	err = archive.Close()
	tu.AssertNoErr(t, err)
	dir := os.TempDir()
	path := filepath.Join(dir, "archive.zip")
	err = os.WriteFile(path, content.Bytes(), os.ModePerm)
	tu.AssertNoErr(t, err)
	t.Log("Written in", path)

	err = api.remove(i1, i2)
	tu.AssertNoErr(t, err)

	i1, i2, err = api.createFile("other.png", tu.PngData)
	tu.AssertNoErr(t, err)

	err = api.remove(i1, i2)
	tu.AssertNoErr(t, err)
}

func TestFilesAPI(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	_, api, _ := devEnv(t)

	err := api.createFile(ca.ComptaFile{Id: 1, Name: "test.png"}, tu.PngData)
	tu.AssertNoErr(t, err)

	err = api.createFile(ca.ComptaFile{Id: 2, Name: "test.png"}, tu.PngData)
	tu.AssertNoErr(t, err)

	b, err := api.download(1, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, bytes.Equal(tu.PngData, b))

	b, err = api.download(2, true)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(b) > 100)

	var content bytes.Buffer
	archive := zip.NewWriter(&content)
	err = api.downloadToZip([]idAndName{{1, "premier"}, {2, "deuxième"}}, ca.ComptaFiles{1: {}, 2: {}}, archive)
	tu.AssertNoErr(t, err)
	err = archive.Close()
	tu.AssertNoErr(t, err)
	tu.Write(t, "archive.zip", content.Bytes())

	err = api.remove(1)
	tu.AssertNoErr(t, err)

	err = api.remove(2)
	tu.AssertNoErr(t, err)

	err = api.createFile(ca.ComptaFile{Id: 3, Name: "other.png"}, tu.PngData)
	tu.AssertNoErr(t, err)

	err = api.remove(3)
	tu.AssertNoErr(t, err)

	files, err := ca.SelectAllComptaFiles(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 0)
}
