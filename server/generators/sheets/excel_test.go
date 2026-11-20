package sheets

import (
	"bytes"
	"testing"
	"time"

	"registro/logic"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"

	"github.com/xuri/excelize/v2"
)

func TestStyle(t *testing.T) {
	m := map[Style]int{}
	m[Style{}] = 5 // check for crashes

	b := NewBuilder()
	_, err := b.register(Style{})
	tu.AssertNoErr(t, err)
}

func TestCreateTable(t *testing.T) {
	liste := [][]Cell{
		{{Value: "lmkeke", Bold: true}, {Value: "Blue", Color: "#AA00BB"}},
		{{ValueF: 5.56, NumFormat: FrancsSuisse}, {ValueF: 3, NumFormat: Int}},
		{{ValueF: 5.56, NumFormat: Euros}, {ValueF: 3, NumFormat: Euros}},
		{{Value: "5.56"}, {ValueF: 0.255, NumFormat: Percentage}},
		{{NumFormat: Float}, {}},
	}
	content, err := CreateTable([]string{"nom", "Prénom"}, liste)
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_Table.xlsx", content)

	content, err = CreateTableTotal([]string{"nom", "Prénom"}, liste, "123.5€")
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_TableTotal.xlsx", content)
}

func TestCreateComplex(t *testing.T) {
	// TODO: consolidate with real data

	content, err := SuiviFinancierCamp([][]Cell{
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
	}, "1354.5€", "1546.4€")
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_SuiviFinancierCamp.xlsx", content)
}

func TestFormatTime(t *testing.T) {
	for _, test := range []struct {
		t        time.Time
		expected string
	}{
		{time.Time{}, ""},
		{time.Date(2000, time.January, 3, 1, 1, 12, 0, time.Local), "03/01/2000 01:01:12"},
	} {
		tu.Assert(t, formatTime(test.t) == test.expected)
	}
}

func TestListeParticipants(t *testing.T) {
	camp := cps.Camp{DateDebut: shared.NewDateFrom(time.Now())}
	p1 := pr.Personne{Identite: pr.Identite{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Man,
		DateNaissance: shared.NewDate(2000, time.August, 5),
		Nationnalite:  pr.Nationnalite{IsSuisse: true},
	}}
	p2 := p1
	p2.Nationnalite.IsSuisse = false
	inscrit1 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 1, IdDossier: 1},
		Personne:    p1,
	}
	inscrit2 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 2, IdDossier: 2},
		Personne:    p2,
	}
	g1, g2 := cps.Groupe{Nom: "Groupe 1", Couleur: "#AA12BB"}, cps.Groupe{Nom: "Groupe 2"}
	liste := []cps.ParticipantPersonne{inscrit1, inscrit2, inscrit1}
	dossiers := logic.Dossiers{Dossiers: dossiers.Dossiers{
		1: dossiers.Dossier{MomentInscription: time.Now(), IdResponsable: 2},
	}}
	fiches := map[cps.IdParticipant]pr.Fichesanitaire{
		1: {AutreContact: pr.NomTel{Nom: "La grand-mère", Tel: "0682084101"}},
	}
	forms := []Form{
		{
			Nom:    "Formulaire 1",
			Champs: cps.Champs{{Titre: "C1"}, {Titre: "C2"}, {Titre: "C3"}},
			Reponses: map[cps.IdParticipant][]string{
				1: {"R1", "R2", "R3"},
				2: {"", "", ""},
			},
		},
		{
			Nom:    "Formulaire 2",
			Champs: cps.Champs{{Titre: "C1"}, {Titre: "C2"}},
			Reponses: map[cps.IdParticipant][]string{
				1: {"", ""},
				2: {"", ""},
			},
		},
	}

	content, err := ListeParticipantsCamp(camp, liste, dossiers, map[cps.IdParticipant]cps.Groupe{1: g1, 2: g2}, fiches, forms, false)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamp_1.xlsx", content)

	content, err = ListeParticipantsCamp(camp, liste, dossiers, map[cps.IdParticipant]cps.Groupe{1: g1, 2: g2}, fiches, nil, true)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamp_2.xlsx", content)
}

func TestListeParticipantsCamps(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	// init.sql registers the default taux (id 1), supporting Euros only.
	const idTauxEuros dossiers.IdTaux = 1
	tauxMulti, err := dossiers.Taux{Label: "Euros et CHF", Euros: 1000, FrancsSuisse: 1100}.Insert(db)
	tu.AssertNoErr(t, err)

	responsable, err := pr.Personne{Identite: pr.Identite{Nom: "Responsable", Prenom: "Test"}}.Insert(db)
	tu.AssertNoErr(t, err)

	campMulti, err := cps.Camp{
		IdTaux: tauxMulti.Id, Nom: "Multi's", DateDebut: shared.NewDateFrom(time.Now()),
		Duree: 7, Prix: dossiers.NewEuros(100),
	}.Insert(db)
	tu.AssertNoErr(t, err)
	campEuros, err := cps.Camp{
		IdTaux: idTauxEuros, Nom: "Euros", DateDebut: shared.NewDateFrom(time.Now()),
		Duree: 7, Prix: dossiers.NewEuros(100),
	}.Insert(db)
	tu.AssertNoErr(t, err)

	dossierMulti, err := dossiers.Dossier{
		IdTaux: tauxMulti.Id, IdResponsable: responsable.Id,
		MomentInscription: time.Now(), PartageAdressesOK: true,
	}.Insert(db)
	tu.AssertNoErr(t, err)
	dossierEuros, err := dossiers.Dossier{
		IdTaux: idTauxEuros, IdResponsable: responsable.Id,
		MomentInscription: time.Now().Add(time.Hour), PartageAdressesOK: false,
	}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = ds.Paiement{IdDossier: dossierMulti.Id, Montant: ds.Montant{Cent: 5000, Currency: ds.Euros}}.Insert(db.DB)
	tu.AssertNoErr(t, err)
	_, err = ds.Paiement{IdDossier: dossierMulti.Id, Montant: ds.Montant{Cent: 1000, Currency: ds.FrancsSuisse}}.Insert(db.DB)
	tu.AssertNoErr(t, err)

	p1, err := pr.Personne{Identite: pr.Identite{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Man,
		DateNaissance: shared.NewDate(2000, time.August, 5),
		Nationnalite:  pr.Nationnalite{IsSuisse: true},
	}}.Insert(db)
	tu.AssertNoErr(t, err)
	p2, err := pr.Personne{Identite: pr.Identite{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Woman,
		DateNaissance: shared.NewDate(2001, time.June, 12),
		Nationnalite:  pr.Nationnalite{IsSuisse: false},
	}}.Insert(db)
	tu.AssertNoErr(t, err)

	part1, err := cps.Participant{
		IdCamp: campMulti.Id, IdPersonne: p1.Id, IdDossier: dossierMulti.Id,
		IdTaux: tauxMulti.Id, Statut: cps.Inscrit,
	}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{
		IdCamp: campMulti.Id, IdPersonne: p2.Id, IdDossier: dossierMulti.Id,
		IdTaux: tauxMulti.Id, Statut: cps.Inscrit,
	}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{
		IdCamp: campEuros.Id, IdPersonne: p2.Id, IdDossier: dossierEuros.Id,
		IdTaux: idTauxEuros, Statut: cps.Inscrit,
	}.Insert(db)
	tu.AssertNoErr(t, err)

	remisesHints := map[cps.IdParticipant]cps.Remises{
		part1.Id: {Equipiers: 10, Famille: 5},
	}

	finances, err := logic.LoadDossiersFinances(db, dossierMulti.Id, dossierEuros.Id)
	tu.AssertNoErr(t, err)
	dataMulti, dataEuros := finances.For(dossierMulti.Id), finances.For(dossierEuros.Id)
	liste := append(dataMulti.ParticipantsExt(), dataEuros.ParticipantsExt()...)

	t.Run("multi currency", func(t *testing.T) {
		content, err := ListeParticipantsCamps(liste, finances, remisesHints, false)
		tu.AssertNoErr(t, err)
		assertListeParticipantsCamps(t, content, 2,
			[]string{
				"Montant payé (devise du camp)",
				"Montant payé (EUR)",
				"Montant payé (CHF)",
				"Fonds de soutien (devise du camp)",
				"Remises (devise du camp)",
			}, nil)
		tu.Write(t, "ListeParticipantsCamps_multi.xlsx", content)
	})

	t.Run("euros only", func(t *testing.T) {
		financesEuros, err := logic.LoadDossiersFinances(db, dossierEuros.Id)
		tu.AssertNoErr(t, err)
		data := financesEuros.For(dossierEuros.Id)
		listeEuros := data.ParticipantsExt()

		content, err := ListeParticipantsCamps(listeEuros, financesEuros, remisesHints, true)
		tu.AssertNoErr(t, err)
		assertListeParticipantsCamps(t, content, 1,
			[]string{
				"Montant payé (devise du camp)",
				"Fonds de soutien (devise du camp)",
				"Remises (devise du camp)",
			}, []string{
				"Montant payé (EUR)",
				"Montant payé (CHF)",
			})
		tu.Write(t, "ListeParticipantsCamps_euros.xlsx", content)
	})
}

// assertListeParticipantsCamps checks the headers of the generated workbook
// and that the monetary columns of every data row are numeric cells.
func assertListeParticipantsCamps(t *testing.T, content []byte, nbRows int, present, absent []string) {
	t.Helper()

	book, err := excelize.OpenReader(bytes.NewReader(content))
	tu.AssertNoErr(t, err)
	defer book.Close()

	rows, err := book.GetRows("Sheet1")
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(rows) >= nbRows+1)
	tu.Assert(t, rows[0][2] == "Devise du camp")
	tu.Assert(t, rows[1][2] == "€")

	headers := make(map[string]int, len(rows[0]))
	for col, header := range rows[0] {
		headers[header] = col + 1
	}
	for _, header := range present {
		col, exists := headers[header]
		tu.Assert(t, exists)
		for row := 2; row <= nbRows+1; row++ {
			cell, err := excelize.CoordinatesToCellName(col, row)
			tu.AssertNoErr(t, err)
			cellType, err := book.GetCellType("Sheet1", cell)
			tu.AssertNoErr(t, err)
			// excelize leaves the type unset for numeric cells written by SetCellValue.
			tu.Assert(t, cellType == excelize.CellTypeNumber || cellType == excelize.CellTypeUnset)
		}
	}
	for _, header := range absent {
		_, exists := headers[header]
		tu.Assert(t, !exists)
	}
}
