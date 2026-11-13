package compta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	cps "registro/sql/camps"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"
)

type migrationFakeDrive struct {
	checkErr    error
	createCalls int
	failOnName  string
	files       map[string][]byte
	removed     []string
}

func newMigrationFakeDrive() *migrationFakeDrive {
	return &migrationFakeDrive{files: map[string][]byte{}}
}

func (f *migrationFakeDrive) check() error { return f.checkErr }

func (f *migrationFakeDrive) createFile(name string, content []byte) (string, string, error) {
	if f.failOnName == name {
		return "", "", errors.New("panne simulée")
	}
	f.createCalls++
	id := fmt.Sprintf("drive-%d", f.createCalls)
	miniatureID := fmt.Sprintf("miniature-%d", f.createCalls)
	f.files[id] = append([]byte(nil), content...)
	f.files[miniatureID] = []byte("miniature")
	return id, miniatureID, nil
}

func (f *migrationFakeDrive) remove(fileID, miniatureID string) error {
	delete(f.files, fileID)
	delete(f.files, miniatureID)
	f.removed = append(f.removed, fileID, miniatureID)
	return nil
}

type migrationFakeLocal struct {
	dir string
}

func (m migrationFakeLocal) paths(id ca.IdComptaFile) (string, string) {
	return filepath.Join(m.dir, fmt.Sprintf("file_%d", id)), filepath.Join(m.dir, fmt.Sprintf("file_%d_min", id))
}

func (m migrationFakeLocal) download(id ca.IdComptaFile, isMiniature bool) ([]byte, error) {
	filePath, miniaturePath := m.paths(id)
	if isMiniature {
		filePath = miniaturePath
	}
	return os.ReadFile(filePath)
}

func (m migrationFakeLocal) write(t *testing.T, id ca.IdComptaFile, content []byte) {
	t.Helper()
	filePath, miniaturePath := m.paths(id)
	tu.AssertNoErr(t, os.WriteFile(filePath, content, os.ModePerm))
	tu.AssertNoErr(t, os.WriteFile(miniaturePath, []byte("miniature"), os.ModePerm))
}

func newMigrationTestDB(t *testing.T) tu.TestDB {
	t.Helper()
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	t.Cleanup(db.Remove)
	return db
}

func newMigrationCamp(t *testing.T, db ca.DB, year int, status ca.FileStatus) (cps.Camp, ca.ComptaCamp) {
	t.Helper()
	camp, err := cps.Camp{
		IdTaux:    1,
		Nom:       fmt.Sprintf("Camp %d", year),
		DateDebut: sh.NewDate(year, time.July, 1),
		Duree:     7,
		Lieu:      "Chamaloc",
	}.Insert(db)
	tu.AssertNoErr(t, err)

	compta, err := ca.ComptaCamp{IdCamp: camp.Id, FileStatus: status}.Insert(db)
	tu.AssertNoErr(t, err)

	return camp, compta
}

func addMigrationFile(t *testing.T, db ca.DB, local migrationFakeLocal, compta ca.ComptaCamp, name string, content []byte) ca.ComptaFile {
	t.Helper()

	meta, err := ca.ComptaFile{Name: name}.Insert(db)
	tu.AssertNoErr(t, err)

	record, err := ca.ComptaRecord{IdCamp: compta.Id, Account: ca.AccountCamp}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = ca.ComptaSubrecord{
		IdRecord: record.Id,
		Date:     sh.NewDateFrom(time.Now()),
		Currency: ds.Euros,
		Montants: ca.Montants{{Kind: ca.Alimentation}},
		IdFile:   ca.NewOptIdComptaFile(meta.Id),
	}.Insert(db)
	tu.AssertNoErr(t, err)

	local.write(t, meta.Id, content)
	return meta
}

func TestMigrateFilesToGoogleDriveDryRun(t *testing.T) {
	db := newMigrationTestDB(t)
	dir := t.TempDir()
	local := migrationFakeLocal{dir: dir}
	drive := newMigrationFakeDrive()

	_, compta := newMigrationCamp(t, db, 2024, ca.FileSystem)
	meta := addMigrationFile(t, db, local, compta, "facture.png", tu.PngData)

	report, err := migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{MaxYear: 2024})
	tu.AssertNoErr(t, err)

	tu.Assert(t, report.CampsSelected == 1)
	tu.Assert(t, report.FilesToUpload == 1)
	tu.Assert(t, report.FilesUploaded == 0)
	tu.Assert(t, drive.createCalls == 0)

	metaAfter, err := ca.SelectComptaFile(db, meta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, metaAfter.DriveID == "")

	comptaAfter, err := ca.SelectComptaCamp(db, compta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, comptaAfter.FileStatus == ca.FileSystem)

	filePath, _ := local.paths(meta.Id)
	tu.Assert(t, fileExists(filePath))
	_, err = os.Stat(report.TrashDir)
	tu.Assert(t, os.IsNotExist(err))
}

func TestMigrateFilesToGoogleDriveApply(t *testing.T) {
	db := newMigrationTestDB(t)
	dir := t.TempDir()
	local := migrationFakeLocal{dir: dir}
	drive := newMigrationFakeDrive()

	_, targetCompta := newMigrationCamp(t, db, 2023, ca.FileSystem)
	f1 := addMigrationFile(t, db, local, targetCompta, "facture.png", tu.PngData)
	f2 := addMigrationFile(t, db, local, targetCompta, "recu.png", append([]byte(nil), tu.PngData...))

	_, futureCompta := newMigrationCamp(t, db, 2025, ca.FileSystem) // après l'année choisie
	newMigrationCamp(t, db, 2022, ca.GoogleDrive)                   // déjà migré
	newMigrationCamp(t, db, 2022, ca.Archived)                      // archivé

	report, err := migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{
		MaxYear: 2024,
		Apply:   true,
		Confirm: func(MigrationReport) bool { return true },
	})
	tu.AssertNoErr(t, err)

	tu.Assert(t, !report.HasFailures())
	tu.Assert(t, len(report.CampsMigrated) == 1 && report.CampsMigrated[0] == targetCompta.Id)
	tu.Assert(t, report.FilesUploaded == 2)
	tu.Assert(t, report.FilesMovedToTrash == 2)
	tu.Assert(t, drive.createCalls == 2)

	for _, id := range []ca.IdComptaFile{f1.Id, f2.Id} {
		meta, err := ca.SelectComptaFile(db, id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, meta.DriveID != "" && meta.DriveIDMiniature != "")
		tu.Assert(t, drive.files[meta.DriveID] != nil)

		filePath, miniaturePath := local.paths(id)
		tu.Assert(t, !fileExists(filePath))
		tu.Assert(t, !fileExists(miniaturePath))
	}

	comptaAfter, err := ca.SelectComptaCamp(db, targetCompta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, comptaAfter.FileStatus == ca.GoogleDrive)

	futureAfter, err := ca.SelectComptaCamp(db, futureCompta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, futureAfter.FileStatus == ca.FileSystem)

	campTrash := filepath.Join(report.TrashDir, fmt.Sprintf("camp_%d", targetCompta.Id))
	tu.Assert(t, fileExists(filepath.Join(campTrash, fmt.Sprintf("file_%d", f1.Id))))
	tu.Assert(t, fileExists(filepath.Join(campTrash, fmt.Sprintf("file_%d_min", f1.Id))))
	tu.Assert(t, fileExists(filepath.Join(campTrash, "manifest.json")))
}

func TestMigrateFilesToGoogleDriveSkipsAlreadyMigratedCamps(t *testing.T) {
	db := newMigrationTestDB(t)
	dir := t.TempDir()
	local := migrationFakeLocal{dir: dir}
	drive := newMigrationFakeDrive()

	_, compta := newMigrationCamp(t, db, 2024, ca.GoogleDrive)
	meta := addMigrationFile(t, db, local, compta, "deja.png", tu.PngData)

	meta.DriveID, meta.DriveIDMiniature = "existing-drive", "existing-miniature"
	_, err := meta.Update(db)
	tu.AssertNoErr(t, err)

	report, err := migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{
		MaxYear: 2024,
		Apply:   true,
		Confirm: func(MigrationReport) bool { return true },
	})
	tu.AssertNoErr(t, err)

	tu.Assert(t, !report.HasFailures())
	tu.Assert(t, report.CampsSelected == 0)
	tu.Assert(t, report.CampsSkipped == 1)
	tu.Assert(t, report.FilesToUpload == 0)
	tu.Assert(t, drive.createCalls == 0)

	metaAfter, err := ca.SelectComptaFile(db, meta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, metaAfter.DriveID == "existing-drive")

	filePath, _ := local.paths(meta.Id)
	tu.Assert(t, fileExists(filePath))
	_, err = os.Stat(report.TrashDir)
	tu.Assert(t, os.IsNotExist(err))
}

func TestMigrateFilesToGoogleDriveFailureRollsBackCamp(t *testing.T) {
	db := newMigrationTestDB(t)
	dir := t.TempDir()
	local := migrationFakeLocal{dir: dir}
	drive := newMigrationFakeDrive()
	drive.failOnName = "seconde.png"

	_, compta := newMigrationCamp(t, db, 2024, ca.FileSystem)
	first := addMigrationFile(t, db, local, compta, "premiere.png", tu.PngData)
	second := addMigrationFile(t, db, local, compta, "seconde.png", tu.PngData)

	report, err := migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{
		MaxYear: 2024,
		Apply:   true,
		Confirm: func(MigrationReport) bool { return true },
	})
	tu.AssertNoErr(t, err)

	tu.Assert(t, report.HasFailures())
	tu.Assert(t, report.FilesUploaded == 0)
	tu.Assert(t, report.FilesMovedToTrash == 0)

	comptaAfter, err := ca.SelectComptaCamp(db, compta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, comptaAfter.FileStatus == ca.FileSystem)

	// The transaction was rolled back: the file uploaded before the failure
	// has no DriveID and its Drive files were removed.
	firstAfter, err := ca.SelectComptaFile(db, first.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, firstAfter.DriveID == "")
	tu.Assert(t, len(drive.files) == 0)

	for _, id := range []ca.IdComptaFile{first.Id, second.Id} {
		filePath, miniaturePath := local.paths(id)
		tu.Assert(t, fileExists(filePath))
		tu.Assert(t, fileExists(miniaturePath))
	}

	// A second run, once the Drive issue is fixed, migrates the whole camp.
	drive.failOnName = ""
	callsBefore := drive.createCalls
	report, err = migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{
		MaxYear: 2024,
		Apply:   true,
		Confirm: func(MigrationReport) bool { return true },
	})
	tu.AssertNoErr(t, err)

	tu.Assert(t, !report.HasFailures())
	tu.Assert(t, report.FilesUploaded == 2)
	tu.Assert(t, drive.createCalls == callsBefore+2)

	comptaAfter, err = ca.SelectComptaCamp(db, compta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, comptaAfter.FileStatus == ca.GoogleDrive)
}

func TestMigrateFilesToGoogleDriveConfirmationAborts(t *testing.T) {
	db := newMigrationTestDB(t)
	dir := t.TempDir()
	local := migrationFakeLocal{dir: dir}
	drive := newMigrationFakeDrive()

	_, compta := newMigrationCamp(t, db, 2024, ca.FileSystem)
	meta := addMigrationFile(t, db, local, compta, "facture.png", tu.PngData)

	_, err := migrateFilesToGoogleDrive(db.DB, drive, local, dir, MigrationOptions{
		MaxYear: 2024,
		Apply:   true,
		Confirm: func(MigrationReport) bool { return false },
	})
	tu.Assert(t, errors.Is(err, ErrMigrationAborted))
	tu.Assert(t, drive.createCalls == 0)

	metaAfter, err := ca.SelectComptaFile(db, meta.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, metaAfter.DriveID == "")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
