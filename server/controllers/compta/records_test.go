package compta

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	ca "registro/sql/compta"
	"registro/sql/shared"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"

	"github.com/benoitkugler/pdf/reader"
)

func TestRecords(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	ct := Controller{db: db.DB}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	l, err := ct.getRecords(co.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l.Equipiers) == 1) // un directeur
	tu.Assert(t, len(l.Records) == 0)

	rec1, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountAsso})
	tu.AssertNoErr(t, err)
	rec2, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountCamp})
	tu.AssertNoErr(t, err)

	l, err = ct.getRecords(co.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l.Records) == 2)

	newDate := time.Date(2002, 1, 12, 0, 0, 0, 0, time.UTC)
	rec1.Subrecords[0].Montants = ca.Montants{{Kind: ca.Activites}, {Kind: ca.Activites}, {Kind: ca.Camping}}
	rec1.Subrecords[0].Date = shared.NewDateFrom(newDate)
	rec1.Record.IsValidated = true
	updated, err := ct.updateRecord(rec1, userAdmin)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Time(updated.Subrecords[0].Date).Equal(newDate))
	tu.Assert(t, updated.Record.IsValidated)

	_, err = ct.uploadJustificatif(rec1.Subrecords[0].Id, []justificatifUpload{
		{filename: "test.png", content: tu.PngData},
	}, userAdmin)
	tu.AssertNoErr(t, err)

	_, _, err = ct.getJustificatif(rec1.Subrecords[0].Id, userAdmin, true)
	tu.AssertNoErr(t, err)

	_, fn, err := ct.getJustificatif(rec1.Subrecords[0].Id, userAdmin, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, fn == "test.png")

	_, err = ct.uploadJustificatif(rec1.Subrecords[0].Id, []justificatifUpload{
		{filename: "first.png", content: pngTestImage(2, 3)},
		{filename: "second.png", content: pngTestImage(5, 4)},
	}, userAdmin)
	tu.AssertNoErr(t, err)
	merged, fn, err := ct.getJustificatif(rec1.Subrecords[0].Id, userAdmin, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, fn == "justificatifs.pdf")
	doc, _, err := reader.ParsePDFReader(bytes.NewReader(merged), reader.Options{})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(doc.Catalog.Pages.Flatten()) == 2)

	_, _, err = ct.getRecordsExcel(co.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.removeJustificatif(rec1.Subrecords[0].Id, userAdmin)
	tu.AssertNoErr(t, err)

	err = ct.deleteSubrecord(rec1.Subrecords[0].Id, userAdmin)
	tu.AssertNoErr(t, err)
	err = ct.deleteSubrecord(rec2.Subrecords[0].Id, userAdmin)
	tu.AssertNoErr(t, err)

	time.Sleep(100 * time.Millisecond) // make sure the delete calls are done
}

func TestMerge(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	ct := Controller{db: db.DB}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	rec1, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountAsso})
	tu.AssertNoErr(t, err)
	rec2, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountCamp})
	tu.AssertNoErr(t, err)

	_, err = ct.addSubrecord(rec1.Record.Id, userAdmin)
	tu.AssertNoErr(t, err)
	_, err = ct.addSubrecord(rec2.Record.Id, userAdmin)
	tu.AssertNoErr(t, err)
	_, err = ct.addSubrecord(rec2.Record.Id, userAdmin)
	tu.AssertNoErr(t, err)

	out, err := ct.mergeRecordTo(MergeRecordToIn{rec2.Record.Id, rec1.Record.Id}, userAdmin)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Subrecords) == 2+3)
}

func TestReviews(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	_, _, smtp := devEnv(t)
	ct := Controller{db: db.DB, smtp: smtp, adminNotificationMail: "bench26@gmail.com"}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	rec1, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountCamp})
	tu.AssertNoErr(t, err)

	_, err = ct.addRecordReview(AddRecordReviewIn{IdRecord: rec1.Record.Id, Text: "Premier message"}, userAdmin)
	tu.AssertNoErr(t, err)

	out, err := ct.addRecordReview(AddRecordReviewIn{IdRecord: rec1.Record.Id, Text: "Réponse"}, co.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Reviews) == 2)
}

func TestSQLDate(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	ct := Controller{db: db.DB}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	group, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountAsso})
	tu.AssertNoErr(t, err)

	dateUTC := time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC)
	var dateLocale sh.Date
	err = json.Unmarshal([]byte(`"2024-08-07"`), &dateLocale)
	tu.AssertNoErr(t, err)

	group.Subrecords[0].Date = sh.Date(dateUTC)
	group.Subrecords[0].Montants = ca.Montants{{Kind: ca.Alimentation}}
	u1, err := ct.updateRecord(group, userAdmin)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Time(u1.Subrecords[0].Date).Day() == 7)

	group.Subrecords[0].Date = dateLocale
	u2, err := ct.updateRecord(group, userAdmin)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Time(u2.Subrecords[0].Date).Day() == 7)
}

func TestUploadAndCreate(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	google, local, _ := devEnv(t)

	ct := Controller{db: db.DB, files: filesAPI{google, local}}
	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.uploadAndCreate(co.Id, []justificatifUpload{
		{filename: "test.png", content: tu.PngData},
	}, co.Id)
	tu.AssertNoErr(t, err)
}
