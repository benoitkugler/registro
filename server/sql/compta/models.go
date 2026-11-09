// Package compta provides the types required
// to register financial operations operated during camps.
package compta

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	"registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/shared"
)

type (
	IdComptaCamp      int64
	IdComptaRecord    int64
	IdComptaSubrecord int64
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

	// Fields only relevant for ACVE

	Account  AccountACVE
	Fraismdr FraismdrACVE
}

// ComptaRecord is one financial operation, which may
// consist of several sub-records (for instance for refunds to equipiers)
//
// gomacro:SQL ADD UNIQUE(Id, IdCamp)
type ComptaRecord struct {
	Id     IdComptaRecord
	IdCamp IdComptaCamp `gomacro-sql-on-delete:"CASCADE"`

	// IsReceipe is [true] if this operation was credited on the association bank.
	IsReceipe bool

	// the expense was made with private funds, and the asso
	// must refund the equipier
	SpentBy OptIdEquipier `gomacro-sql-foreign:"Equipier" `

	Reviews     Reviews
	IsValidated bool // rapprochement and validation by the admin

	Account AccountKind // the default value is OK for Repère
}

type ComptaSubrecord struct {
	Id       IdComptaSubrecord
	IdRecord IdComptaRecord `gomacro-sql-on-delete:"CASCADE"`

	Date        shared.Date
	Description string
	Mode        ModePaiementRecord // ignored for [ComptaRecord.Account] == [AccountAsso]
	Serial      string             // mandatory for [ComptaRecord.Mode] Cheque, ignored otherwise

	// [Currency] is constant for one given record
	Currency ds.Currency
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
// gomacro:SQL ADD UNIQUE(IdSubrecord)
type ComptaRecordBus struct {
	IdCamp      IdComptaCamp      `gomacro-sql-on-delete:"CASCADE"`
	IdSubrecord IdComptaSubrecord `gomacro-sql-on-delete:"CASCADE"`
}

// ComptaBugdet is an 'extension' table for a camp,
// storing the estimated budget.
//
// gomacro:SQL ADD UNIQUE(IdCompta)
type ComptaBugdet struct {
	IdCompta IdComptaCamp `gomacro-sql-on-delete:"CASCADE"`

	// the following fields may be set to forecast a budget
	// before the real inscriptions are open
	SimulatedInscrits  int
	SimulatedEquipiers int

	StartMealInscrits, EndMealInscrits               Meal
	StartMealEquipiers, EndMealEquipiers             Meal
	OffsetStartMealEquipiers, OffsetEndMealEquipiers int

	ExternalHost Rents // ignored for camp at MDR

	Activities Activities

	MaterielPeda  ds.Montant
	ProduitPharma ds.Montant

	TeamTrips       ds.Montant
	ExternalCarRent ds.Montant
	AssoCarKms      CarsKms
}

// ComptaConfig holds parameters used to compute the budget,
// which may change from one year to another.
// For now, we support (at most) one config per year
//
// gomacro:SQL ADD UNIQUE(Year)
type ComptaConfig struct {
	Id   IdComptaConfig
	Year int

	MealsCost          MealsCosts
	SmallMealThreshold int
	// additional percentage (typically 10) added to the meals
	// if the number of participants is strictly lower than the [SmallMealThreshold]
	SmallMealMajoration int

	AssoCars CarsCosts

	// ACVE only, in €

	ChauffageMDR Cent // cost for one day at MDR
	EntretienMDR Cent // cost for one day and one participant or equipier at MDR
	MaterielMDR  Cent // cost for one day and one participant at MDR
}
