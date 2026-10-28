// Package compta provides the types required
// to register financial operations operated during camps.
package compta

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	"registro/sql/camps"
	"registro/sql/shared"
)

type (
	IdComptaCamp      int64
	IdComptaBankentry int64
	IdComptaRecord    int64
	IdComptaConfig    int64
	IdComptaFile      int64
)

// ComptaCamp is an extension table adding
// compta features to one camp.
//
// gomacro:SQL ADD UNIQUE(IdCamp)
type ComptaCamp struct {
	Id     IdComptaCamp
	IdCamp camps.IdCamp `gomacro-sql-on-delete:"CASCADE"`

	FileStatus FileStatus

	State State // see also [State] enum

	LastSentMailAdmin, LastSentMailDirecteur time.Time // used to avoid spamming mails during a review
}

// ComptaCampACVE add informations only relevant for ACVE
type ComptaCampACVE struct {
	IdComptaCamp IdComptaCamp `gomacro-sql-on-delete:"CASCADE"`

	Account  AccountACVE
	Fraismdr FraismdrACVE
}

// ComptaBankentry is one financial operation, which may
// consist of several records (for instance for refunds to equipiers)
//
// gomacro:SQL ADD UNIQUE(Id, IdCamp)
type ComptaBankentry struct {
	Id     IdComptaBankentry
	IdCamp IdComptaCamp `gomacro-sql-on-delete:"CASCADE"`

	// IsReceipe is [true] if this operation was credited on the association bank.
	IsReceipe bool

	Reviews     Reviews
	IsValidated bool // rapprochement and validation by the admin

	Account AccountKind // the default value is OK for Repère
}

type ComptaRecord struct {
	Id          IdComptaRecord
	IdBankentry IdComptaBankentry `gomacro-sql-on-delete:"CASCADE"`

	Date        shared.Date
	Description string
	Mode        ModePaiementRecord // ignored for [ComptaBankentry.Account] == [AccountAsso]
	Serial      string             // mandatory for [ComptaRecord.Mode] Cheque, ignored otherwise

	Montants Montants // avec les ventilations

	IdFile OptIdComptaFile `gomacro-sql-foreign:"ComptaFile" gomacro-sql-on-delete:"SET NULL"`
}

// ComptaFile stores metadata for uploaded files
// To save storage, files are either
//   - directly stored on the (server) file system : fast but expensive
//   - stored on a Google drive : free but slow (limited as well)
//   - archived (on an external storage), not accessible anymore : for older files
type ComptaFile struct {
	Id               IdComptaFile
	Name             string
	DriveID          string
	DriveIDMiniature string
}

// ComptaRecordBus is a link table marking a record as used for
// bus expenses.
// For such records, the ventilations related to cars will be deduced from the kilometric indemnity
//
// gomacro:SQL ADD UNIQUE(IdRecord)
type ComptaRecordBus struct {
	IdCamp   IdComptaCamp   `gomacro-sql-on-delete:"CASCADE"`
	IdRecord IdComptaRecord `gomacro-sql-on-delete:"CASCADE"`
}

// TODO
// // ComptaBugdet is an 'extension' table for a camp,
// // storing the estimated budget.
// //
// // gomacro:SQL ADD UNIQUE(IdCompta)
// type ComptaBugdet struct {
// 	IdCompta IdComptaCamp `gomacro-sql-on-delete:"CASCADE"`

// 	BonusTeamPeople int // additional people present during the vacation
// 	MealsInscrits   MealsCount
// 	MealsEquipiers  MealsCount

// 	ExternalHost RentCosts

// 	Activities    Activities
// 	MaterielPeda  Cent
// 	ProduitPharma Cent

// 	TeamTrips       Cent
// 	ExternalCarRent Cent

// 	// the following fields may be set to forecast a budget
// 	// before the real inscriptions are open
// 	SimulatedInscrits  int
// 	SimulatedEquipiers int
// 	SimulatedReduc1    int // number of participants with one parent in the team
// 	SimulatedReduc2    int // number of participants with two parents in the team

// 	CarsMDRKms CarsKms
// }

// // Comptaconfig holds parameters used to compute the budget,
// // which may change from one year to another.
// // For now, we support (at most) one config per year
// //
// // gomacro:SQL ADD UNIQUE(Year)
// type Comptaconfig struct {
// 	Id   IdComptaConfig
// 	Year int

// 	MealsCost          MealsCosts
// 	SmallMealThreshold int
// 	// additional percentage (typically 10) added to the meals
// 	// if the number of participants is strictly lower than the [SmallMealThreshold]
// 	SmallMealMajoration int

// 	ChauffageMDR Cent // cost for one day at MDR
// 	EntretienMDR Cent // cost for one day and one participant or equipier at MDR
// 	MaterielMDR  Cent // cost for one day and one participant at MDR

// 	CarsMDR CarsCosts
// }
