// Package compta implements a shared document
// to register financial operations operated during camps.
package compta

import "time"

type (
	IdCompta       int64
	IdRecord       int64
	IdSubrecord    int64
	IdDonc         int64
	IdBudgetconfig int64
	IdFilecompta   int64
)

// gomacro:SQL ADD UNIQUE(IdCamp)
type Compta struct {
	Id          IdCompta
	IdCamp      int64 `gomacro-sql-foreign:"Camp" gomacro-sql-on-delete:"CASCADE"`
	Account     Account
	RecordNotes string

	Fraismdr Fraismdr

	FileStatus FileStatus

	State State // see also [State] enum

	LastReviewNotif time.Time
}

// Record is one financial operation
//
// gomacro:SQL ADD UNIQUE(Id, IdCompta)
type Record struct {
	Id       IdRecord
	IdCompta IdCompta `gomacro-sql-on-delete:"CASCADE"`

	IsOtherAccount bool // false if in the main account
	IsReceipe      bool

	Reviews     Reviews
	IsValidated bool // rapprochement and validation by the admin
}

type Subrecord struct {
	Id       IdSubrecord
	IdRecord IdRecord `gomacro-sql-on-delete:"CASCADE"`

	Date   Date
	Title  string
	Mode   Mode   // ignored for OtherAccount
	Serial string // mandatory for Cheque, ignored otherwise

	Montants Montants

	IdFilecompta OptionnalIdFilecompta `gomacro-sql-foreign:"Filecompta" gomacro-sql-on-delete:"SET NULL"`
}

// Filecompta stores metadata for uploaded files
// To save storage, files are either
//   - directly stored on the (server) file system : fast but expensive
//   - stored on a Google drive : free but slow (limited as well)
//   - archived (on an external storage), not accessible anymore : for older files
type Filecompta struct {
	Id               IdFilecompta
	Name             string
	DriveID          string
	DriveIDMiniature string
}

type Mode uint8

const (
	CarteB    Mode = iota // Carte bancaire
	Virement              // Virement
	Liquide               // Liquide
	Cheque                // Chèque
	Helloasso             // Helloasso
)

func (m Mode) String() string {
	switch m {
	case CarteB:
		return "Carte bancaire"
	case Virement:
		return "Virement"
	case Liquide:
		return "Liquide"
	case Cheque:
		return "Chèque"
	case Helloasso:
		return "Helloasso"
	default:
		return "<invalid mode>"
	}
}

// Donc is a donation for a camp
type Donc struct {
	Id       IdDonc
	IdCompta IdCompta `gomacro-sql-on-delete:"CASCADE"`

	Date    Date
	Title   string
	Mode    Mode
	Serial  string // mandatory for Cheque, ignored otherwise
	Montant Cent
}

// ComptaBudget is an 'extension' table for a camp,
// storing the estimated budget.
//
// gomacro:SQL ADD UNIQUE(IdCompta)
type ComptaBudget struct {
	IdCompta IdCompta `gomacro-sql-on-delete:"CASCADE"`

	BonusTeamPeople int // additional people present during the vacation
	MealsInscrits   MealsCount
	MealsEquipiers  MealsCount

	ExternalHost RentCosts

	Activities    Activities
	MaterielPeda  Cent
	ProduitPharma Cent

	TeamTrips       Cent
	ExternalCarRent Cent

	// the following fields may be set to forecast a budget
	// before the real inscriptions are open
	SimulatedInscrits  int
	SimulatedEquipiers int
	SimulatedReduc1    int // number of participants with one parent in the team
	SimulatedReduc2    int // number of participants with two parents in the team

	CarsMDRKms CarsKms
}

// Budgetconfig holds parameters used to compute the budget,
// which may change from one year to another.
// For now, we support (at most) one config per year
//
// gomacro:SQL ADD UNIQUE(Year)
type Budgetconfig struct {
	Id   IdBudgetconfig
	Year int

	MealsCost          MealsCosts
	SmallMealThreshold int
	// additional percentage (typically 10) added to the meals
	// if the number of participants is strictly lower than the [SmallMealThreshold]
	SmallMealMajoration int

	ChauffageMDR Cent // cost for one day at MDR
	EntretienMDR Cent // cost for one day and one participant or equipier at MDR
	MaterielMDR  Cent // cost for one day and one participant at MDR

	CarsMDR CarsCosts
}

// RecordBus is a link table marking a record as used for
// bus expenses.
// For such records, the ventilations related to cars will be deduced from the kilometric indemnity
//
// gomacro:SQL ADD UNIQUE(IdSubrecord)
type RecordBus struct {
	IdCompta    IdCompta    `gomacro-sql-on-delete:"CASCADE"`
	IdSubrecord IdSubrecord `gomacro-sql-on-delete:"CASCADE"`
}
