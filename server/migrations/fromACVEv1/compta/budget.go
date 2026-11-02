package compta

import (
	"registro/migrations/fromACVEv1/rawdata"
)

// ----------------------------------- Budget -----------------------------------

// Number of meals during the whole vacation, for ONE person
type MealsCount [5]int32 // petit déjeuner, midi, gouter, soir cinquième

type RentCost struct {
	Name string
	Cost Cent
}

type RentCosts []RentCost

type Activity struct {
	Name string
	Cost Cent
}

type Activities []Activity

// MealsCosts stores the cost, in cent, of one meal
type MealsCosts [5]Cent

// ResolvedBudget exposes the dynamic fields, computed
// using camp statistics
type ResolvedBudget struct {
	Duree      int // days
	Inscrits   int
	Equipiers  int // including additional team members
	DaysPeople int // number of days * number of people
	Prix       rawdata.Euros

	FoodInscrits       float32
	FoodEquipiers      float32
	HasSmallTeamFactor bool
	FoodTotal          float32

	ChauffageMDR float32
	EntretienMDR float32
	HostTotal    float32

	MaterielMDR     float32
	ActivitiesTotal float32

	TransportTotal float32

	ReductionsTotal float32

	PensionsNet     float32 // pensions - reductions
	Depenses        float32
	Bilan           float32
	BilanPercentage int
	AmountAccount   float32
}
