package compta

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"registro/config"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	"registro/utils"
)

// ErrMigrationAborted is returned when the user declines the confirmation
// before an actual migration run.
var ErrMigrationAborted = errors.New("migration annulée")

// MigrationOptions configures [MigrateFilesToGoogleDrive].
type MigrationOptions struct {
	// MaxYear selects the camps whose start date is in this year or before.
	MaxYear int
	// Apply is false for a dry run (no upload, no database update, no move).
	Apply bool
	// Confirm is called in apply mode, after the plan has been computed and
	// before any change is made. Returning false aborts the migration.
	// It may be nil, in which case the migration starts immediately.
	Confirm func(MigrationReport) bool
}

// CampFailure describes a camp which could not be migrated.
type CampFailure struct {
	IdCamp   cps.IdCamp
	IdCompta ca.IdComptaCamp
	Err      error
}

// MigrationReport summarizes a (possibly dry) migration run.
type MigrationReport struct {
	CampsScanned  int
	CampsSelected int
	CampsSkipped  int
	CampsMigrated []ca.IdComptaCamp

	FilesToUpload     int
	FilesUploaded     int
	FilesMovedToTrash int
	MissingLocalFiles []ca.IdComptaFile

	TrashDir string
	Failures []CampFailure
	Warnings []string
}

// HasFailures reports whether at least one camp failed to migrate.
func (r MigrationReport) HasFailures() bool { return len(r.Failures) > 0 }

// driveStorage is the subset of [googleDriveAPI] used by the migration.
type driveStorage interface {
	check() error
	createFile(fileName string, content []byte) (string, string, error)
	remove(fileId, miniatureId string) error
}

// localStorage is the subset of [fileSystemAPI] used by the migration.
type localStorage interface {
	download(id ca.IdComptaFile, isMiniature bool) ([]byte, error)
	paths(id ca.IdComptaFile) (filePath, miniaturePath string)
}

type campTarget struct {
	camp   cps.Camp
	compta ca.ComptaCamp
	files  ca.ComptaFiles
}

// MigrateFilesToGoogleDrive moves the compta files of all camps starting
// in [MigrationOptions.MaxYear] or before from the local file system to
// Google Drive.
//
// The migration is idempotent: files already having a DriveID are skipped,
// and a camp status is only switched to [ca.GoogleDrive] once all its files
// are on Drive. Successful local files are moved to a timestamped trash
// folder under COMPTA_FILES_DIR.
//
// Each camp is migrated atomically: the DriveID updates and the camp status
// update are committed in a single transaction. If any step fails, the
// transaction is rolled back and the Drive files uploaded during the attempt
// are removed.
func MigrateFilesToGoogleDrive(db *sql.DB, cfg config.Compta, opts MigrationOptions) (MigrationReport, error) {
	drive := googleDriveAPI{KeyFile: cfg.GoogleDriveKeyFile, FolderId: cfg.GoogleDriveFolderId}
	local := fileSystemAPI{Folder: cfg.FilesDir}
	return migrateFilesToGoogleDrive(db, drive, local, cfg.FilesDir, opts)
}

func migrateFilesToGoogleDrive(db *sql.DB, drive driveStorage, local localStorage, filesDir string, opts MigrationOptions) (MigrationReport, error) {
	var report MigrationReport

	if opts.MaxYear < 2000 {
		return report, fmt.Errorf("année invalide: %d", opts.MaxYear)
	}

	if err := drive.check(); err != nil {
		return report, fmt.Errorf("connexion Google Drive: %w", err)
	}

	targets, err := collectMigrationTargets(db, opts.MaxYear, &report)
	if err != nil {
		return report, err
	}

	report.TrashDir = newTrashDirPath(filesDir)

	log.Printf("%d camp(s) à migrer, %d fichier(s) à envoyer",
		report.CampsSelected, report.FilesToUpload)
	log.Printf("destination des fichiers locaux : %s", report.TrashDir)

	if !opts.Apply {
		report.MissingLocalFiles = findMissingLocalFiles(targets, local)
		for _, id := range report.MissingLocalFiles {
			log.Printf("fichier local manquant : %d", id)
		}
		return report, nil
	}

	if len(targets) == 0 {
		return report, nil
	}

	if opts.Confirm != nil && !opts.Confirm(report) {
		return report, ErrMigrationAborted
	}

	if err := os.MkdirAll(report.TrashDir, os.ModePerm); err != nil {
		return report, fmt.Errorf("création de la corbeille %s: %w", report.TrashDir, err)
	}

	for _, target := range targets {
		err := migrateCamp(db, drive, local, target, report.TrashDir, &report)
		if err != nil {
			report.Failures = append(report.Failures, CampFailure{
				IdCamp:   target.compta.IdCamp,
				IdCompta: target.compta.Id,
				Err:      err,
			})
			log.Printf("camp %d (%s) : échec : %v", target.compta.Id, target.camp.Nom, err)
			continue
		}

		report.CampsMigrated = append(report.CampsMigrated, target.compta.Id)
		log.Printf("camp %d (%s) : migré", target.compta.Id, target.camp.Nom)
	}

	return report, nil
}

func collectMigrationTargets(db ca.DB, maxYear int, report *MigrationReport) ([]campTarget, error) {
	camps, err := cps.SelectAllCamps(db)
	if err != nil {
		return nil, fmt.Errorf("chargement des camps: %w", err)
	}

	comptas, err := ca.SelectAllComptaCamps(db)
	if err != nil {
		return nil, fmt.Errorf("chargement des comptabilités: %w", err)
	}
	byCamp := comptas.ByIdCamp()

	report.CampsScanned = len(camps)

	var targets []campTarget
	for _, camp := range camps {
		compta, hasCompta := byCamp[camp.Id]
		if !hasCompta {
			report.CampsSkipped++
			continue
		}
		if camp.DateDebut.Time().Year() > maxYear {
			report.CampsSkipped++
			continue
		}
		if compta.FileStatus != ca.FileSystem {
			report.CampsSkipped++
			continue
		}

		_, subrecords, err := getRecords(db, compta.Id)
		if err != nil {
			return nil, fmt.Errorf("camp %d: chargement des écritures: %w", camp.Id, err)
		}

		files, err := ca.SelectComptaFiles(db, subrecords.IdFiles()...)
		if err != nil {
			return nil, fmt.Errorf("camp %d: chargement des fichiers: %w", camp.Id, err)
		}

		report.FilesToUpload += len(files)

		targets = append(targets, campTarget{camp: camp, compta: compta, files: files})
	}

	report.CampsSelected = len(targets)

	// Deterministic order: older camps first.
	sort.Slice(targets, func(i, j int) bool {
		di, dj := targets[i].camp.DateDebut.Time(), targets[j].camp.DateDebut.Time()
		if di.Equal(dj) {
			return targets[i].camp.Id < targets[j].camp.Id
		}
		return di.Before(dj)
	})

	return targets, nil
}

func migrateCamp(db *sql.DB, drive driveStorage, local localStorage, target campTarget, trashDir string, report *MigrationReport) error {
	type createdFile struct {
		id          string
		miniatureID string
	}

	var created []createdFile
	uploaded := 0

	err := utils.InTx(db, func(tx *sql.Tx) error {
		for _, id := range sortedFileIDs(target.files) {
			meta := target.files[id]

			content, err := local.download(id, false)
			if err != nil {
				return fmt.Errorf("lecture du fichier %d (%s): %w", id, meta.Name, err)
			}

			driveID, miniatureID, err := drive.createFile(meta.Name, content)
			if err != nil {
				return fmt.Errorf("envoi du fichier %d (%s): %w", id, meta.Name, err)
			}
			created = append(created, createdFile{id: driveID, miniatureID: miniatureID})

			meta.DriveID, meta.DriveIDMiniature = driveID, miniatureID
			updated, err := meta.Update(tx)
			if err != nil {
				return fmt.Errorf("mise à jour SQL du fichier %d (%s): %w", id, meta.Name, err)
			}

			target.files[id] = updated
			uploaded++
		}

		target.compta.FileStatus = ca.GoogleDrive
		if _, err := target.compta.Update(tx); err != nil {
			return fmt.Errorf("mise à jour du statut du camp %d: %w", target.compta.Id, err)
		}

		return nil
	})
	if err != nil {
		// The transaction was rolled back: remove the Drive files uploaded
		// during this attempt so that a retry starts from a clean state.
		for _, file := range created {
			if rmErr := drive.remove(file.id, file.miniatureID); rmErr != nil {
				msg := fmt.Sprintf("camp %d: impossible de supprimer les fichiers Drive orphelins %s/%s: %v", target.compta.Id, file.id, file.miniatureID, rmErr)
				report.Warnings = append(report.Warnings, msg)
				log.Printf("avertissement : %s", msg)
			}
		}
		return err
	}

	report.FilesUploaded += uploaded
	moveCampFilesToTrash(local, target, trashDir, report)
	return nil
}

type trashManifest struct {
	CreatedAt time.Time
	IdCamp    cps.IdCamp
	IdCompta  ca.IdComptaCamp
	Files     []ca.ComptaFile
}

func moveCampFilesToTrash(local localStorage, target campTarget, trashDir string, report *MigrationReport) {
	campDir := filepath.Join(trashDir, fmt.Sprintf("camp_%d", target.compta.Id))
	manifest := trashManifest{
		CreatedAt: time.Now(),
		IdCamp:    target.camp.Id,
		IdCompta:  target.compta.Id,
	}

	moved := false
	for _, id := range sortedFileIDs(target.files) {
		meta := target.files[id]
		manifest.Files = append(manifest.Files, meta)

		filePath, miniaturePath := local.paths(id)
		for i, src := range []string{filePath, miniaturePath} {
			if src == "" {
				continue
			}
			if _, err := os.Stat(src); err != nil {
				if !os.IsNotExist(err) {
					log.Printf("avertissement : inspection de %s impossible : %v", src, err)
				}
				continue
			}

			if err := os.MkdirAll(campDir, os.ModePerm); err != nil {
				log.Printf("avertissement : création de %s impossible : %v", campDir, err)
				return
			}

			dst := filepath.Join(campDir, filepath.Base(src))
			if err := os.Rename(src, dst); err != nil {
				msg := fmt.Sprintf("déplacement de %s vers %s impossible : %v", src, dst, err)
				report.Warnings = append(report.Warnings, msg)
				log.Printf("avertissement : %s", msg)
				continue
			}

			moved = true
			if i == 0 {
				report.FilesMovedToTrash++
			}
		}
	}

	if !moved {
		return
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Printf("avertissement : sérialisation du manifeste de %s impossible : %v", campDir, err)
		return
	}
	if err := os.WriteFile(filepath.Join(campDir, "manifest.json"), content, 0o644); err != nil {
		log.Printf("avertissement : écriture du manifeste de %s impossible : %v", campDir, err)
	}
}

func sortedFileIDs(files ca.ComptaFiles) []ca.IdComptaFile {
	ids := files.IDs()
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func findMissingLocalFiles(targets []campTarget, local localStorage) []ca.IdComptaFile {
	var missing []ca.IdComptaFile
	for _, target := range targets {
		for _, id := range sortedFileIDs(target.files) {
			if target.files[id].DriveID != "" {
				continue
			}
			filePath, _ := local.paths(id)
			if _, err := os.Stat(filePath); err != nil {
				missing = append(missing, id)
			}
		}
	}
	return missing
}

func newTrashDirPath(filesDir string) string {
	base := filepath.Join(filesDir, "trash_compta_drive_"+time.Now().Format("2006-01-02_15-04-05"))
	candidate := base
	for i := 1; ; i++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
}
