package compta

import (
	"testing"
	"time"

	cps "registro/sql/camps"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSelectConfig(t *testing.T) {
	for _, test := range []struct {
		configYears []int
		wantYear    int
		wantError   bool
	}{
		{[]int{2000, 2001, 2023, 2024}, 2023, false},
		{[]int{2000, 2001, 2024}, 2001, false},
		{[]int{2024}, 0, true},
		{[]int{}, 0, true},
		{[]int{2017}, 2017, false},
	} {
		configs := ca.ComptaConfigs{}
		for i, y := range test.configYears {
			configs[ca.IdComptaConfig(i)] = ca.ComptaConfig{Year: y}
		}
		got, err := selectBudgetConfig(2023, configs)
		tu.Assert(t, err != nil == test.wantError)
		tu.Assert(t, got.Year == test.wantYear)
	}
}

func TestResolveMealsCount(t *testing.T) {
	for _, test := range []struct {
		name        string
		startMeal   ca.Meal
		endMeal     ca.Meal
		duree       int
		startOffset int
		endOffset   int
		want        mealsCount
	}{
		{name: "full stay", startMeal: ca.PetitDejeuner, endMeal: ca.Cinquieme, duree: 4, want: mealsCount{4, 4, 4, 4, 4}},
		{name: "middle meals only", startMeal: ca.Midi, endMeal: ca.Soir, duree: 3, want: mealsCount{2, 3, 3, 3, 2}},
		{name: "single meal", startMeal: ca.Gouter, endMeal: ca.Gouter, duree: 2, want: mealsCount{1, 1, 2, 1, 1}},
		{name: "offset modifies duration", startMeal: ca.PetitDejeuner, endMeal: ca.Cinquieme, duree: 3, startOffset: 1, endOffset: 0, want: mealsCount{2, 2, 2, 2, 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := resolveMealsCount(test.startMeal, test.endMeal, test.duree, test.startOffset, test.endOffset)
			tu.Assert(t, got == test.want)
		})
	}
}

func TestResolveBudget(t *testing.T) {
	euros := func(cents int) ds.Montant {
		return ds.Montant{Cent: cents, Currency: ds.Euros}
	}

	for _, test := range []struct {
		name      string
		loader    comptaLoader
		origin    ca.ComptaBugdet
		config    ca.ComptaConfig
		simulated bool
		want      ResolvedBudget
	}{
		{
			name: "simulated small team with external costs",
			loader: comptaLoader{
				camp: cps.Camp{
					Duree:     2,
					DateDebut: shared.NewDate(2026, time.June, 1),
					Lieu:      "Var",
					Prix:      euros(10000),
				},
				taux:         ds.Taux{Euros: 1000},
				participants: cps.Participants{1: {Statut: cps.Inscrit}, 2: {Statut: cps.Inscrit}, 3: {Statut: cps.AttenteCampComplet}},
				equipiers:    cps.Equipiers{1: {}, 2: {}},
			},
			origin: ca.ComptaBugdet{
				SimulatedInscrits:  2,
				SimulatedEquipiers: 1,
				StartMealInscrits:  ca.PetitDejeuner,
				EndMealInscrits:    ca.Cinquieme,
				StartMealEquipiers: ca.PetitDejeuner,
				EndMealEquipiers:   ca.Cinquieme,
				ExternalHost:       ca.Rents{{Name: "hosting", Cost: euros(400)}},
				Activities:         ca.Activities{{Name: "activity", Cost: euros(5000)}},
				MaterielPeda:       euros(2500),
				ProduitPharma:      euros(500),
				TeamTrips:          euros(1000),
				ExternalCarRent:    euros(2000),
				AssoCarKms:         ca.CarsKms{2, 0, 1},
			},
			config: ca.ComptaConfig{
				MealsCost:           ca.MealsCosts{100, 100, 100, 100, 100},
				SmallMealThreshold:  3,
				SmallMealMajoration: 10,
				AssoCars:            ca.CarsCosts{5, 0, 10},
			},
			simulated: true,
			want: ResolvedBudget{
				Duree:              2,
				Inscrits:           2,
				Equipiers:          1,
				Prix:               euros(10000),
				FoodInscrits:       euros(2000),
				FoodEquipiers:      euros(1000),
				HasSmallTeamFactor: true,
				FoodTotal:          euros(3300),
				HostTotal:          euros(400),
				ActivitiesTotal:    euros(8000),
				TransportTotal:     euros(3020),
				Pensions:           euros(20000),
				Depenses:           euros(14720),
				Bilan:              euros(5280),
				BilanPercentage:    26,
			},
		},
		{
			name: "actual team at MDR in winter",
			loader: comptaLoader{
				camp: cps.Camp{
					Duree:     3,
					DateDebut: shared.NewDate(2026, time.January, 1),
					Lieu:      "Chamaloc",
					Prix:      euros(5000),
				},
				taux: ds.Taux{Euros: 1000},
				participants: cps.Participants{
					1: {Statut: cps.Inscrit},
					2: {Statut: cps.Inscrit},
					3: {Statut: cps.Inscrit},
					4: {Statut: cps.AttenteCampComplet},
				},
				equipiers: cps.Equipiers{1: {}, 2: {}},
			},
			origin: ca.ComptaBugdet{
				StartMealInscrits:  ca.PetitDejeuner,
				EndMealInscrits:    ca.Cinquieme,
				StartMealEquipiers: ca.PetitDejeuner,
				EndMealEquipiers:   ca.Cinquieme,
				ExternalHost:       ca.Rents{{Name: "ignored external host", Cost: euros(9000)}},
			},
			config: ca.ComptaConfig{
				MealsCost:           ca.MealsCosts{50, 50, 50, 50, 50},
				SmallMealThreshold:  3,
				SmallMealMajoration: 20,
				ChauffageMDR:        200,
				EntretienMDR:        30,
				MaterielMDR:         10,
			},
			want: ResolvedBudget{
				Duree:           3,
				Inscrits:        3,
				Equipiers:       2,
				Prix:            euros(5000),
				FoodInscrits:    euros(2250),
				FoodEquipiers:   euros(1500),
				FoodTotal:       euros(3750),
				ChauffageMDR:    euros(600),
				EntretienMDR:    euros(450),
				HostTotal:       euros(1050),
				MaterielMDR:     euros(90),
				ActivitiesTotal: euros(90),
				Pensions:        euros(15000),
				Depenses:        euros(4890),
				Bilan:           euros(10110),
				BilanPercentage: 67,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := test.loader.resolveBudget(test.origin, test.config, test.simulated)
			tu.Assert(t, got == test.want)
		})
	}
}
