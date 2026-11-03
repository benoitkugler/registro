package compta

import (
	"fmt"

	"registro/migrations/fromACVEv1/rawdata"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
)

type BudgetPrevisionnelSummary struct {
	Reel   ds.Montant
	Simule ds.Montant
}

// ----------------------------------- Budget -----------------------------------

// Number of meals during the whole vacation, for ONE person
type MealsCount [5]int32 // petit déjeuner, midi, gouter, soir cinquième

type RentCost struct {
	Name string
	Cost Cent
}

type RentCosts []RentCost

func (rc RentCosts) total() Cent {
	var total Cent
	for _, item := range rc {
		total += item.Cost
	}
	return total
}

type Activity struct {
	Name string
	Cost Cent
}

type Activities []Activity

// MealsCosts stores the cost, in cent, of one meal
type MealsCosts [5]Cent

func (mc MealsCosts) resolveCost(people int, meals MealsCount) (total Cent) {
	for i, cost := range mc {
		count := meals[i]
		total += cost * Cent(count)
	}
	return total * Cent(people)
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
	Duree      int // days
	Inscrits   int
	Equipiers  int // including additional team members
	DaysPeople int // number of days * number of people
	Prix       rawdata.Euros

	FoodInscrits       ds.Montant
	FoodEquipiers      ds.Montant
	HasSmallTeamFactor bool
	FoodTotal          ds.Montant

	ChauffageMDR ds.Montant
	EntretienMDR ds.Montant
	HostTotal    ds.Montant

	MaterielMDR     ds.Montant
	ActivitiesTotal ds.Montant

	TransportTotal ds.Montant

	ReductionsTotal ds.Montant

	PensionsNet     ds.Montant // pensions - reductions
	Depenses        ds.Montant
	Bilan           ds.Montant
	BilanPercentage int
	AmountAccount   ds.Montant
}

// if [simulated] is true, the number of inscrits, équipiers, .. are taken from the budget field;
// otherwise they are computed from the inscriptions
func (ld *comptaLoader) resolveBudget(origin ca.ComptaBugdet, config ca.ComptaConfig, simulated bool) ResolvedBudget {
	out := ResolvedBudget{}

	// TODO
	// out.Duree = ld.camp.Duree()
	// if simulated {
	// 	out.Inscrits = origin.SimulatedInscrits
	// 	out.Equipiers = origin.SimulatedEquipiers + origin.BonusTeamPeople
	// } else {
	// 	out.Inscrits = ld.inscrits()
	// 	out.Equipiers = len(ld.equipiers) + origin.BonusTeamPeople
	// }
	// out.DaysPeople = out.Duree * (out.Inscrits + out.Equipiers)
	// out.Prix = ld.camp.Prix

	// out.FoodInscrits = config.MealsCost.resolveCost(out.Inscrits, origin.MealsInscrits).euros()
	// out.FoodEquipiers = config.MealsCost.resolveCost(out.Equipiers, origin.MealsEquipiers).euros()
	// out.HasSmallTeamFactor = out.Inscrits < config.SmallMealThreshold
	// percent := 0
	// if out.HasSmallTeamFactor {
	// 	percent = config.SmallMealMajoration
	// }
	// out.FoodTotal = (out.FoodInscrits + out.FoodEquipiers) * float32(100+percent) / 100

	// atMDR := ld.atMDR()
	// if atMDR {
	// 	// chauffage is activated from november to april
	// 	m := ld.camp.DateDebut.Time().Month()
	// 	if m >= time.November || m <= time.April {
	// 		out.ChauffageMDR = (config.ChauffageMDR * Cent(out.Duree)).euros()
	// 	}
	// 	out.EntretienMDR = (config.EntretienMDR * Cent(out.DaysPeople)).euros()
	// 	// host cost is only MDR
	// 	out.HostTotal = out.ChauffageMDR + out.EntretienMDR
	// } else {
	// 	// host cost is computed with user provided data
	// 	out.HostTotal = origin.ExternalHost.total().euros()
	// }

	// materielMDR := Cent(0)
	// if atMDR {
	// 	materielMDR = config.MaterielMDR * Cent(out.Inscrits) * Cent(out.Duree)
	// }
	// // total activities
	// var activitiesTotal Cent
	// for _, activity := range origin.Activities {
	// 	activitiesTotal += activity.Cost
	// }
	// activitiesTotal += materielMDR + origin.MaterielPeda + origin.ProduitPharma

	// transportTotal := origin.TeamTrips + origin.ExternalCarRent +
	// 	Cent(origin.CarsMDRKms[CarBus])*config.CarsMDR[CarBus] +
	// 	Cent(origin.CarsMDRKms[CarMinibus])*config.CarsMDR[CarMinibus] +
	// 	Cent(origin.CarsMDRKms[CarCar])*config.CarsMDR[CarCar]

	// var reductionsTotals Cent
	// if simulated {
	// 	reductionsTotals = ld.estimatedReductions(origin.SimulatedReduc1, origin.SimulatedReduc2)
	// } else {
	// 	// use the real number computed from the inscriptions
	// 	reductionsTotals = newCent(ld.camp.Prix)*Cent(out.Inscrits) - ld.computeRecette()
	// }

	// pensionsNet := newCent(out.Prix)*Cent(out.Inscrits) - reductionsTotals

	// out.MaterielMDR = materielMDR.euros()
	// out.ActivitiesTotal = activitiesTotal.euros()
	// out.TransportTotal = transportTotal.euros()
	// out.ReductionsTotal = reductionsTotals.euros()
	// out.PensionsNet = pensionsNet.euros()
	// out.Depenses = out.FoodTotal + out.HostTotal + out.ActivitiesTotal + out.TransportTotal
	// out.Bilan = out.PensionsNet - out.Depenses
	// if out.PensionsNet != 0 {
	// 	out.BilanPercentage = int(100 * out.Bilan / out.PensionsNet)
	// }

	// out.AmountAccount = out.Depenses

	return out
}

// compute the (approximate) rules for team children reductions
//   - > one week : 25/50%
//   - one week : 10/20%
//   - < camp a la journée : 5/10%
// func (ld *comptaLoader) estimatedReductions(reducOne, reducTwo int) Cent {
// 	var p1, p2 Cent // percentage
// 	if ld.camp.OptionPrix.Active == rawdata.OptionsPrix.JOUR {
// 		p1, p2 = 5, 10
// 	} else if ld.camp.Duree() <= 7 {
// 		p1, p2 = 10, 20
// 	} else {
// 		p1, p2 = 25, 50
// 	}
// 	base := newCent(ld.camp.Prix)
// 	return Cent(reducOne)*base*p1/100 + Cent(reducTwo)*base*p2/100
// }
