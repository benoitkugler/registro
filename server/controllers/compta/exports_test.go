package compta

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"registro/controllers/files"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestCampBilanConvertsAmountsToEuros(t *testing.T) {
	taux := ds.Taux{Euros: 1000, FrancsSuisse: 950}
	summary1 := CampSummary{
		Nom:                   "Camp 1",
		Prix:                  ds.Montant{Cent: 1000, Currency: ds.FrancsSuisse},
		Recettes:              ds.Montant{Cent: 2000, Currency: ds.FrancsSuisse},
		Remises:               ds.Montant{Cent: 100, Currency: ds.FrancsSuisse},
		BudgetPrevisionnel:    BudgetPrevisionnelSummary{Reel: ds.Montant{Cent: 300, Currency: ds.FrancsSuisse}},
		DepensesTotal:         ds.Montant{Cent: 500, Currency: ds.FrancsSuisse},
		DepenseJourneeInscrit: ds.Montant{Cent: 50, Currency: ds.FrancsSuisse},
		ventilations:          ventilations{ca.Activites: {accCampDepenses: ds.Montant{Cent: 500, Currency: ds.FrancsSuisse}.Multi()}},
	}

	got := newCampBilan(summary1, taux)
	tu.Assert(t, got.Prix == 950)
	tu.Assert(t, got.Recettes == 1900)
	tu.Assert(t, got.Remises == 95)
	tu.Assert(t, got.DepensesTotal == 475)
	tu.Assert(t, got.DepenseJourneeInscrit == 48)
	tu.Assert(t, got.ventilations[ca.Activites].camp == 475)

	total := (bilanCamps{got}).total()
	tu.Assert(t, total.Recettes == got.Recettes)
	tu.Assert(t, total.DepensesTotal == got.DepensesTotal)
}

func TestGenerateExcelCamps(t *testing.T) {
	camps := bilanCamps{
		{Nom: "Camp 1", Serial: "S1", Directeur: "Alice Martin", Lieu: "Aix-en-Provence", Account: ca.MaisonFamiliale, DateDebut: sh.NewDate(2024, time.January, 10), Duree: 1, JoursInscrits: 1, Inscrits: 1, Equipiers: 2, AuPairs: 1, Prix: 12000, PrixJournee: 12000, Recettes: 10000, DepensesTotal: 3000, DepenseJourneeInscrit: 3000},
		{Nom: "Camp 2", Serial: "S2", Directeur: "Bruno Leroy", Lieu: "Annecy", Account: ca.Clairiere, DateDebut: sh.NewDate(2024, time.February, 15), Duree: 2, JoursInscrits: 4, Inscrits: 2, Equipiers: 3, AuPairs: 1, Prix: 15000, PrixJournee: 7500, Recettes: 20000, DepensesTotal: 1700, DepenseJourneeInscrit: 425},
		{Nom: "Camp 3", Serial: "S3", Directeur: "Claire Dubois", Lieu: "Bordeaux", Account: ca.Rocher, DateDebut: sh.NewDate(2024, time.March, 5), Duree: 3, JoursInscrits: 9, Inscrits: 3, Equipiers: 4, AuPairs: 2, Prix: 18000, PrixJournee: 6000, Recettes: 30000, DepensesTotal: 3700, DepenseJourneeInscrit: 411},
		{Nom: "Camp 4", Serial: "S4", Directeur: "David Morel", Lieu: "Chamonix", Account: ca.Montagne, DateDebut: sh.NewDate(2024, time.April, 12), Duree: 4, JoursInscrits: 16, Inscrits: 4, Equipiers: 5, AuPairs: 2, Prix: 22000, PrixJournee: 5500, Recettes: 40000, DepensesTotal: 2800, DepenseJourneeInscrit: 175},
		{Nom: "Camp 5", Serial: "S5", Directeur: "Élodie Petit", Lieu: "Dijon", Account: ca.Bergerie, DateDebut: sh.NewDate(2024, time.May, 20), Duree: 5, JoursInscrits: 25, Inscrits: 5, Equipiers: 6, AuPairs: 3, Prix: 26000, PrixJournee: 5200, Recettes: 50000, DepensesTotal: 1900, DepenseJourneeInscrit: 76},
	}
	camps[0].ventilations[ca.Activites].camp = 1000
	camps[0].ventilations[ca.Alimentation].asso = 2000
	camps[1].ventilations[ca.Alimentation].camp = 1200
	camps[1].ventilations[ca.Activites].asso = 500
	camps[2].ventilations[ca.Hotelier].camp = 3000
	camps[2].ventilations[ca.Carburant].asso = 700
	camps[3].ventilations[ca.Bank].camp = 800
	camps[3].ventilations[ca.Hotelier].asso = 2000
	camps[4].ventilations[ca.Activites].camp = 400
	camps[4].ventilations[ca.Alimentation].camp = 600
	camps[4].ventilations[ca.Carburant].asso = 900

	buf, err := generateExcelCamps(camps)
	tu.AssertNoErr(t, err)
	tu.Write(t, "bilan_camps.xlsx", buf.Bytes())
}

func TestGenerateExcelCampsWithFraisMDR(t *testing.T) {
	camps := bilanCamps{
		{Nom: "Camp MDR", Serial: "S-MDR", Directeur: "Alice Martin", Lieu: "Chamaloc", Account: ca.MaisonFamiliale, DateDebut: sh.NewDate(2024, time.January, 12), Duree: 7, JoursInscrits: 21, Inscrits: 3, Equipiers: 4, AuPairs: 1, Prix: 18000, PrixJournee: 2571, Recettes: 22000, DepensesTotal: 13000, DepenseJourneeInscrit: 619, fraisMDRVehicule: 2500, fraisMDRMateriel: 1800},
		{Nom: "Camp sans MDR", Serial: "S-STD", Directeur: "Bruno Leroy", Lieu: "Annecy", Account: ca.Clairiere, DateDebut: sh.NewDate(2024, time.February, 18), Duree: 3, JoursInscrits: 9, Inscrits: 3, Equipiers: 3, AuPairs: 1, Prix: 15000, PrixJournee: 5000, Recettes: 17000, DepensesTotal: 9000, DepenseJourneeInscrit: 333},
	}
	camps[0].ventilations[ca.Activites].camp = 3500
	camps[0].ventilations[ca.Alimentation].asso = 1200
	camps[0].ventilations[ca.MaintenanceVehic].camp = 900
	camps[1].ventilations[ca.Hotelier].camp = 2800
	camps[1].ventilations[ca.Carburant].asso = 700

	buf, err := generateExcelCamps(camps)
	tu.AssertNoErr(t, err)
	tu.Write(t, "bilan_camps_frais_mdr.xlsx", buf.Bytes())
}

func TestGenerateExcelCampsWithCampAccountVentilations(t *testing.T) {
	camps := bilanCamps{
		{Nom: "Camp compte dédié", Serial: "S-CAMP", Account: ca.MaisonFamiliale, Duree: 2, Inscrits: 1, JoursInscrits: 2},
		{Nom: "Camp compte dédié 2", Serial: "S-CAMP", Account: ca.MaisonFamiliale, Duree: 2, Inscrits: 1, JoursInscrits: 2},
	}
	camps[0].ventilations[ca.Activites].camp = 1500
	camps[0].ventilations[ca.Alimentation].camp = 800
	camps[1].ventilations[ca.Activites].camp = 800

	tu.Assert(t, !camps.showAssoAccount())

	buf, err := generateExcelCamps(camps)
	tu.AssertNoErr(t, err)
	tu.Write(t, "bilan_camps_no_asso.xlsx", buf.Bytes())
}

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
