package compta

import (
	"fmt"

	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
)

type BudgetPrevisionnelSummary struct {
	Reel   ds.Montant
	Simule ds.Montant
}

// returns the more recent configuration among the ones available at the sejour time
func selectBudgetConfig(campYear int, configs ca.ComptaConfigs) (ca.ComptaConfig, error) {
	sorted := configs.Sorted()
	// find the best match, that is the first year before (or equal) [campYear]
	for _, config := range sorted {
		if config.Year <= campYear { // got it
			return config, nil
		}
	}
	return ca.ComptaConfig{}, fmt.Errorf("internal error: aucune configuration n'est disponible pour l'année %d", campYear)
}

// ResolvedBudget exposes the dynamic fields, computed
// using camp statistics
type ResolvedBudget struct {
	Duree     int // days
	Inscrits  int
	Equipiers int
	Prix      ds.Montant

	FoodInscrits       ds.Montant
	FoodEquipiers      ds.Montant
	HasSmallTeamFactor bool
	FoodTotal          ds.Montant

	ChauffageMDR ds.Montant
	EntretienMDR ds.Montant
	HostTotal    ds.Montant // external or MDR (if any)

	MaterielMDR     ds.Montant
	ActivitiesTotal ds.Montant

	TransportTotal ds.Montant

	Pensions        ds.Montant
	Depenses        ds.Montant
	Bilan           ds.Montant // pensions - depenses
	BilanPercentage int
}

// if [simulated] is true, the number of inscrits and equipiers are taken from the budget field
// otherwise they are computed from the database
func (ld *comptaLoader) resolveBudget(origin ca.ComptaBugdet, config ca.ComptaConfig, simulated bool) ResolvedBudget {
	out := ResolvedBudget{}

	out.Duree = ld.camp.Duree
	if simulated {
		out.Inscrits = origin.SimulatedInscrits
		out.Equipiers = origin.SimulatedEquipiers
	} else {
		out.Inscrits = ld.inscrits()
		out.Equipiers = len(ld.equipiers)
	}
	out.Prix = ld.camp.Prix

	foodInscrits := resolveMealsCost(config.MealsCost, out.Inscrits, resolveMealsCount(origin.StartMealInscrits, origin.EndMealInscrits, out.Duree, 0, 0))
	foodEquipiers := resolveMealsCost(config.MealsCost, out.Equipiers, resolveMealsCount(origin.StartMealEquipiers, origin.EndMealEquipiers, out.Duree, origin.OffsetStartMealEquipiers, origin.OffsetEndMealEquipiers))
	out.HasSmallTeamFactor = out.Inscrits < config.SmallMealThreshold
	percent := 0
	if out.HasSmallTeamFactor {
		percent = config.SmallMealMajoration
	}
	out.FoodInscrits = ds.Montant{Cent: foodInscrits, Currency: ds.Euros}
	out.FoodEquipiers = ds.Montant{Cent: foodEquipiers, Currency: ds.Euros}
	out.FoodTotal = ds.Montant{Cent: (foodInscrits + foodEquipiers) * (100 + percent) / 100, Currency: ds.Euros}

	atMDR := ld.atMDR()
	if atMDR {
		out.ChauffageMDR, out.EntretienMDR = ld.budgetFraisMDR(config, out.Inscrits+out.Equipiers)
		// host cost is only MDR
		out.HostTotal = ds.Montant{Cent: out.ChauffageMDR.Cent + out.EntretienMDR.Cent, Currency: ds.Euros}
	} else {
		// host cost is computed with user provided data
		out.HostTotal = resolveCostList(ld.taux, origin.ExternalHost)
	}

	var materielMDR ds.Montant
	if atMDR {
		materielMDR = ds.Montant{Cent: config.MaterielMDR * out.Inscrits * out.Duree, Currency: ds.Euros}
	}
	// total activities
	activitiesTotal := ld.taux.Zero()
	activitiesTotal.Add(materielMDR)
	activitiesTotal.Add(resolveCostList(ld.taux, origin.Activities))
	activitiesTotal.Add(origin.MaterielPeda)
	activitiesTotal.Add(origin.ProduitPharma)

	transportTotal := ld.taux.Zero()
	transportTotal.Add(origin.TeamTrips)
	transportTotal.Add(origin.ExternalCarRent)
	transportTotal.Add(origin.AssoCarKms.Resolve(config.AssoCars))

	out.MaterielMDR = materielMDR
	out.ActivitiesTotal = activitiesTotal.Montant
	out.TransportTotal = transportTotal.Montant

	depenses := ld.taux.Zero()
	depenses.Add(out.FoodTotal)
	depenses.Add(out.HostTotal)
	depenses.Add(out.ActivitiesTotal)
	depenses.Add(out.TransportTotal)

	out.Pensions = out.Prix.Mult(out.Inscrits)
	out.Depenses = depenses.Montant
	depensesC := ld.taux.Convertible(out.Depenses).Convert(out.Pensions.Currency)
	out.Bilan = ds.Montant{Cent: out.Pensions.Cent - depensesC.Cent, Currency: out.Pensions.Currency}
	if out.Pensions.Cent != 0 {
		out.BilanPercentage = 100 * out.Bilan.Cent / out.Pensions.Cent
	}

	return out
}

// Number of meals during the whole vacation, for ONE person
type mealsCount [ca.Cinquieme + 1]int

func resolveMealsCount(startMeal, endMeal ca.Meal, duree, startOffset, endOffset int) mealsCount {
	// start by full days, then adjust
	duree += (endOffset - startOffset)
	out := mealsCount{duree, duree, duree, duree, duree}
	for i := range startMeal { // remove one meal before start (if any)
		out[i] -= 1
	}
	for i := endMeal + 1; i < ca.Cinquieme+1; i++ { // remove one meal after end (if any)
		out[i] -= 1
	}
	return out
}

// in € cents
func resolveMealsCost(mc ca.MealsCosts, people int, meals mealsCount) Cent {
	var total Cent
	for i, cost := range mc {
		count := meals[i]
		total += Cent(cost) * Cent(count)
	}
	return total * people
}

func resolveCostList(taux ds.Taux, costs []ca.NamedCost) ds.Montant {
	out := taux.Zero()
	for _, c := range costs {
		out.Add(c.Cost)
	}
	return out.Montant
}
