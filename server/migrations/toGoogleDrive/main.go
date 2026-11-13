// Script de migration des justificatifs compta du disque local vers Google Drive.
//
// Il migre les camps dont la date de début est dans l'année choisie ou avant.
// Le script est en mode simulation par défaut ; utiliser -apply pour écrire.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"registro/config"
	ct "registro/controllers/compta"
)

func main() {
	year := flag.Int("year", 0, "migre les camps dont la date de début est dans cette année ou avant (requis)")
	apply := flag.Bool("apply", false, "effectue réellement la migration (par défaut : simulation)")
	flag.Parse()

	if *year < 2000 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go -year <année> [-apply]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	dbCfg, err := config.NewDB()
	check(err)

	db, err := dbCfg.ConnectPostgres()
	check(err)
	defer db.Close()

	if err := db.Ping(); err != nil {
		check(fmt.Errorf("connexion à la base: %w", err))
	}

	comptaCfg, err := config.NewCompta()
	check(err)

	opts := ct.MigrationOptions{
		MaxYear: *year,
		Apply:   *apply,
	}
	if *apply {
		opts.Confirm = confirm
	}

	report, err := ct.MigrateFilesToGoogleDrive(db, comptaCfg, opts)
	if err != nil {
		if errors.Is(err, ct.ErrMigrationAborted) {
			fmt.Println("Migration annulée.")
			return
		}
		log.Fatal(err)
	}

	printReport(report, *apply)

	if report.HasFailures() {
		os.Exit(1)
	}
}

func confirm(report ct.MigrationReport) bool {
	fmt.Printf("\n%d camp(s), %d fichier(s) à envoyer vers Google Drive.\n", report.CampsSelected, report.FilesToUpload)
	if len(report.MissingLocalFiles) > 0 {
		fmt.Printf("Attention : %d fichier(s) local(aux) manquant(s).\n", len(report.MissingLocalFiles))
	}
	fmt.Printf("Les fichiers locaux migrés seront déplacés vers :\n  %s\n", report.TrashDir)
	fmt.Print("Continuer ? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes", "o", "oui":
		return true
	default:
		return false
	}
}

func printReport(report ct.MigrationReport, apply bool) {
	if !apply {
		fmt.Println("\nSimulation terminée, aucune modification effectuée.")
	} else {
		fmt.Println("\nMigration terminée.")
	}

	fmt.Printf("Camps examinés : %d\n", report.CampsScanned)
	fmt.Printf("Camps sélectionnés : %d\n", report.CampsSelected)
	fmt.Printf("Camps ignorés : %d\n", report.CampsSkipped)
	fmt.Printf("Camps migrés : %d\n", len(report.CampsMigrated))
	fmt.Printf("Fichiers à envoyer : %d\n", report.FilesToUpload)
	fmt.Printf("Fichiers envoyés : %d\n", report.FilesUploaded)
	fmt.Printf("Fichiers déplacés vers la corbeille : %d\n", report.FilesMovedToTrash)

	if report.TrashDir != "" {
		fmt.Printf("Corbeille : %s\n", report.TrashDir)
	}

	if len(report.MissingLocalFiles) > 0 {
		fmt.Println("\nFichiers locaux manquants :")
		for _, id := range report.MissingLocalFiles {
			fmt.Printf("  - fichier %d\n", id)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\nAvertissements :")
		for _, warning := range report.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if len(report.Failures) > 0 {
		fmt.Println("\nÉchecs :")
		for _, failure := range report.Failures {
			fmt.Printf("  - camp %d (compta %d) : %v\n", failure.IdCamp, failure.IdCompta, failure.Err)
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
