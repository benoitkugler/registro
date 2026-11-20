package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"

	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/compta"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

func TestTrimParticipants(t *testing.T) {
	participants := rd.Participants{
		1: {Id: 1, IdCamp: 10, IdFacture: rd.OptionnalId{Int64: 100, Valid: true}},
		2: {Id: 2, IdCamp: 20, IdFacture: rd.OptionnalId{Int64: 200, Valid: true}},
		3: {Id: 3, IdCamp: 10},
	}
	factures := rd.Factures{
		100: {Id: 100},
		200: {Id: 200},
	}

	gotParticipants, gotFactures := trimParticipants(participants, factures, baseV1{camps: rd.Camps{10: {Id: 10}}})
	if len(gotParticipants) != 2 || gotParticipants[1].Id != 1 || gotParticipants[3].Id != 3 {
		t.Fatalf("trimmed participants = %+v, want participants 1 and 3", gotParticipants)
	}
	if len(gotFactures) != 1 || gotFactures[100].Id != 100 {
		t.Fatalf("trimmed factures = %+v, want facture 100 only", gotFactures)
	}
}

func TestTrimPersonnesKeepsRegistrationPeople(t *testing.T) {
	trimmed := baseV1{
		factures: rd.Factures{100: {Id: 100, IdPersonne: 1}},
		participants: rd.Participants{
			2: {Id: 2, IdPersonne: 2},
		},
	}
	got := trimPersonnes(rd.Personnes{
		1: {Id: 1},
		2: {Id: 2},
		3: {Id: 3},
	}, trimmed)
	if len(got) != 2 || got[1].Id != 1 || got[2].Id != 2 {
		t.Errorf("trimPersonnes() = %+v, want dossier owner and participant only", got)
	}
}

func TestDossierFromV1(t *testing.T) {
	moment := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.UTC)
	got := dossierFromV1(rd.Facture{
		Key:               "legacy-key",
		CopiesMails:       pq.StringArray{"copy@example.org"},
		PartageAdressesOK: true,
	}, 1, 42, moment)

	want := ds.Dossier{
		IdTaux:            1,
		IdResponsable:     42,
		CopiesMails:       pr.Mails{"copy@example.org"},
		PartageAdressesOK: true,
		MomentInscription: moment,
		KeyV1:             "legacy-key",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dossierFromV1() = %+v, want %+v", got, want)
	}
}

func TestFactureMomentsInscription(t *testing.T) {
	first := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.UTC)
	last := first.Add(24 * time.Hour)
	got := factureMomentsInscription(rd.Participants{
		1: {IdFacture: rd.NewOptionnalId(100), DateHeure: rd.Time(last)},
		2: {IdFacture: rd.NewOptionnalId(100), DateHeure: rd.Time{}},
		3: {IdFacture: rd.NewOptionnalId(100), DateHeure: rd.Time(first)},
	})
	if got[100] != first {
		t.Errorf("factureMomentsInscription() = %v, want earliest time %v", got[100], first)
	}
}

func TestParticipantConversions(t *testing.T) {
	status, err := statutParticipant(rd.ListeAttente{Statut: rd.AttenteReponse})
	if err != nil || status != cps.EnAttenteReponse {
		t.Fatalf("statutParticipant() = %v, %v; want %v", status, err, cps.EnAttenteReponse)
	}

	gotRemises := remises(rd.Remises{
		ReducEquipiers: 10,
		ReducEnfants:   20,
		ReducSpeciale:  12.34,
	})
	wantRemises := cps.Remises{
		Equipiers: 10,
		Famille:   20,
		Speciale:  ds.Montant{Cent: 1234, Currency: ds.Euros},
	}
	if gotRemises != wantRemises {
		t.Errorf("remises() = %+v, want %+v", gotRemises, wantRemises)
	}

	gotOption := optionPrixParticipant(rd.OptionPrixParticipant{Jour: rd.Jours{1, 3}})
	wantOption := cps.OptionPrixParticipant{Jour: cps.JoursPresence{1, 3}}
	if !reflect.DeepEqual(gotOption, wantOption) {
		t.Errorf("optionPrixParticipant() = %+v, want %+v", gotOption, wantOption)
	}
}

func TestParticipantFromV1(t *testing.T) {
	got, err := participantFromV1(rd.Participant{
		ListeAttente:     rd.ListeAttente{Statut: rd.Inscrit},
		Remises:          rd.Remises{ReducEquipiers: 5, ReducEnfants: 10, ReducSpeciale: 7.5},
		QuotientFamilial: 420,
		OptionPrix:       rd.OptionPrixParticipant{Jour: rd.Jours{0, 2}},
	}, 11, 22, 33, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := cps.Participant{
		IdCamp:           11,
		IdPersonne:       22,
		IdDossier:        33,
		IdTaux:           1,
		Statut:           cps.Inscrit,
		Remises:          cps.Remises{Equipiers: 5, Famille: 10, Speciale: ds.Montant{Cent: 750, Currency: ds.Euros}},
		QuotientFamilial: 420,
		OptionPrix:       cps.OptionPrixParticipant{Jour: cps.JoursPresence{0, 2}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("participantFromV1() = %+v, want %+v", got, want)
	}
}

func TestPresenceOffsets(t *testing.T) {
	camp := shared.Plage{From: shared.NewDate(2025, time.July, 10), Duree: 5}

	t.Run("inactive presence", func(t *testing.T) {
		got := presenceOffsets(rd.OptionnalPlage{
			Plage:  rd.Plage{From: rd.Date(time.Date(2025, time.July, 11, 0, 0, 0, 0, time.UTC))},
			Active: false,
		}, camp)
		if got != (cps.PresenceOffsets{}) {
			t.Errorf("presenceOffsets(inactive) = %+v, want zero offsets", got)
		}
	})

	t.Run("active presence dates are relative to camp", func(t *testing.T) {
		got := presenceOffsets(rd.OptionnalPlage{
			Plage: rd.Plage{
				From: rd.Date(time.Date(2025, time.July, 12, 0, 0, 0, 0, time.UTC)),
				To:   rd.Date(time.Date(2025, time.July, 14, 0, 0, 0, 0, time.UTC)),
			},
			Active: true,
		}, camp)
		want := cps.PresenceOffsets{Debut: 2, Fin: 0}
		if got != want {
			t.Errorf("presenceOffsets(active) = %+v, want %+v", got, want)
		}

		got = presenceOffsets(rd.OptionnalPlage{
			Plage: rd.Plage{
				From: rd.Date(time.Date(2025, time.July, 10, 0, 0, 0, 0, time.UTC)),
				To:   rd.Date(time.Date(2025, time.July, 13, 0, 0, 0, 0, time.UTC)),
			},
			Active: true,
		}, camp)
		want = cps.PresenceOffsets{Debut: 0, Fin: -1}
		if got != want {
			t.Errorf("presenceOffsets(active) = %+v, want %+v", got, want)
		}
	})
}

func TestTrimBudgets(t *testing.T) {
	budgets := comptaV1.ComptaBudgets{
		{IdCompta: 1},
		{IdCompta: 2},
	}
	trimmed := trimBudgets(budgets, baseV1{
		comptas: comptaV1.Comptas{
			1: {},
		},
	})
	if len(trimmed) != 1 || trimmed[0].IdCompta != 1 {
		t.Fatalf("trimBudgets() = %+v, want only budget for compta 1", trimmed)
	}
}

func TestBudgetFromV1(t *testing.T) {
	got := budgetFromV1(comptaV1.ComptaBudget{
		SimulatedInscrits:  12,
		SimulatedEquipiers: 3,
		ExternalHost: comptaV1.RentCosts{
			{Name: "Gite", Cost: 1200},
		},
		Activities: comptaV1.Activities{
			{Name: "Kayak", Cost: 3500},
		},
		MaterielPeda:    400,
		ProduitPharma:   500,
		TeamTrips:       600,
		ExternalCarRent: 700,
		CarsMDRKms:      comptaV1.CarsKms{11, 22, 33},
	}, 42)

	if got.IdCompta != 42 || got.SimulatedInscrits != 12 || got.SimulatedEquipiers != 3 {
		t.Fatalf("budget identity/simulated counts = %+v", got)
	}
	if got.ExternalHost[0] != (compta.NamedCost{Name: "Gite", Cost: ds.Montant{Cent: 1200, Currency: ds.Euros}}) {
		t.Errorf("ExternalHost = %+v", got.ExternalHost)
	}
	if got.Activities[0] != (compta.NamedCost{Name: "Kayak", Cost: ds.Montant{Cent: 3500, Currency: ds.Euros}}) {
		t.Errorf("Activities = %+v", got.Activities)
	}
	if got.MaterielPeda != (ds.Montant{Cent: 400, Currency: ds.Euros}) ||
		got.ProduitPharma != (ds.Montant{Cent: 500, Currency: ds.Euros}) ||
		got.TeamTrips != (ds.Montant{Cent: 600, Currency: ds.Euros}) ||
		got.ExternalCarRent != (ds.Montant{Cent: 700, Currency: ds.Euros}) {
		t.Errorf("budget amounts = %+v", got)
	}
	if got.AssoCarKms != (compta.CarsKms{11, 22, 33}) {
		t.Errorf("AssoCarKms = %v", got.AssoCarKms)
	}
}

func TestBudgetConfigFromV1(t *testing.T) {
	got := budgetConfigFromV1(comptaV1.Budgetconfig{
		Year:                2024,
		MealsCost:           comptaV1.MealsCosts{10, 20, 30, 40, 50},
		SmallMealThreshold:  8,
		SmallMealMajoration: 15,
		ChauffageMDR:        100,
		EntretienMDR:        200,
		MaterielMDR:         300,
		CarsMDR:             comptaV1.CarsCosts{400, 500, 600},
	})

	if got.Year != 2024 || got.SmallMealThreshold != 8 || got.SmallMealMajoration != 15 {
		t.Fatalf("config identity/thresholds = %+v", got)
	}
	if got.MealsCost != (compta.MealsCosts{10, 20, 30, 40, 50}) {
		t.Errorf("MealsCost = %v", got.MealsCost)
	}
	if got.ChauffageMDR != 100 || got.EntretienMDR != 200 || got.MaterielMDR != 300 {
		t.Errorf("MDR costs = %d, %d, %d", got.ChauffageMDR, got.EntretienMDR, got.MaterielMDR)
	}
	if got.AssoCars != (compta.CarsCosts{400, 500, 600}) {
		t.Errorf("AssoCars = %v", got.AssoCars)
	}
}
