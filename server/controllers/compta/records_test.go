package compta

import (
	"bytes"
	"encoding/json"
	ca "registro/sql/compta"
	"registro/sql/shared"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"
	"testing"
	"time"
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
	tu.Assert(t, len(l) == 0)

	rec1, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountAsso})
	tu.AssertNoErr(t, err)
	rec2, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountCamp})
	tu.AssertNoErr(t, err)

	l, err = ct.getRecords(co.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 2)

	newDate := time.Date(2002, 1, 12, 0, 0, 0, 0, time.UTC)
	rec1.Subrecords[0].Montants = ca.Montants{{Kind: ca.Activites}, {Kind: ca.Activites}, {Kind: ca.Camping}}
	rec1.Subrecords[0].Date = shared.NewDateFrom(newDate)
	rec1.Record.IsValidated = true
	updated, err := ct.updateRecord(rec1, userAdmin)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Time(updated.Subrecords[0].Date).Equal(newDate))
	tu.Assert(t, updated.Record.IsValidated)

	_, err = ct.uploadJustificatif(rec1.Subrecords[0].Id, "test.png", bytes.NewReader(tu.PngData), userAdmin)
	tu.AssertNoErr(t, err)

	_, _, err = ct.getJustificatif(rec1.Subrecords[0].Id, userAdmin, true)
	tu.AssertNoErr(t, err)

	_, fn, err := ct.getJustificatif(rec1.Subrecords[0].Id, userAdmin, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, fn == "test.png")

	_, _, err = ct.getRecordsArchive(co.Id)
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

	_, err = ct.addRecordReview(AddRecordReviewIn{IdBankentry: rec1.Record.Id, Text: "Premier message"}, userAdmin)
	tu.AssertNoErr(t, err)

	out, err := ct.addRecordReview(AddRecordReviewIn{IdBankentry: rec1.Record.Id, Text: "Réponse"}, co.Id)
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
	group.Subrecords[0].Montants[0].Kind = ca.Alimentation
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

	_, err = ct.uploadAndCreate(co.Id, "test.png", bytes.NewReader(tu.PngData), co.Id)
	tu.AssertNoErr(t, err)
}
