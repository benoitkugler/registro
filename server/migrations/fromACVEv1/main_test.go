package main

import (
	"testing"
	"time"

	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/compta"
	ds "registro/sql/dossiers"
	"registro/sql/shared"
)

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
