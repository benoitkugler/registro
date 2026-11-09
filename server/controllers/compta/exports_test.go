package compta

import (
	"bytes"
	"fmt"
	"testing"

	"registro/controllers/files"
	ca "registro/sql/compta"
	tu "registro/utils/testutils"
)

func TestStreamRecordsArchive(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, _ := createCamps(t, db)
	ct := Controller{
		db:    db.DB,
		files: filesAPI{fileSystem: fileSystemAPI{Folder: t.TempDir()}},
	}

	compta, err := ct.loadOrCreateCompta(camp.Id)
	tu.AssertNoErr(t, err)
	record, err := ct.createRecord(CreateRecordIn{compta.Id, ca.AccountAsso})
	tu.AssertNoErr(t, err)
	_, err = ct.uploadJustificatif(record.Subrecords[0].Id, []justificatifUpload{
		{filename: "receipt.png", content: tu.PngData},
	}, userAdmin)
	tu.AssertNoErr(t, err)

	archiveName, items, err := ct.streamRecordsArchive(compta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, archiveName == fmt.Sprintf("Compta %s.zip", camp.Label()))

	var got []files.ZipItem
	for item, err := range items {
		tu.AssertNoErr(t, err)
		got = append(got, item)
	}
	tu.Assert(t, len(got) == 2)
	tu.Assert(t, got[0].Name == "compta.xlsx" && len(got[0].Content) != 0)
	tu.Assert(t, got[1].Name == fmt.Sprintf("justificatif_%d.png", record.Subrecords[0].Id))
	tu.Assert(t, bytes.Equal(got[1].Content, tu.PngData))
}
