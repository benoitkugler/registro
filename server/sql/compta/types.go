package compta

import (
	"time"

	"registro/sql/camps"
	"registro/sql/dossiers"
	ds "registro/sql/dossiers"
	"registro/sql/shared"
)

// State is the global state of the compta. document
type State uint8

const (
	// The director is editing the document
	EnCours State = iota // Dossier en cours d'édition
	// The director has finished declaring entries,
	// and waits for the compta validation
	ToBeValidated // En attente de validation
	// The compta has validated every entry : every
	// new change will revert the state back to EnCours
	Valide // Dossier validé par la compta
	// The compta has closed the folder : no more changes are allowed
	// by directors
	Archive // Dossier archivé
)

func (st State) String() string {
	switch st {
	case EnCours:
		return "Dossier en cours d'édition"
	case ToBeValidated:
		return "En attente de validation"
	case Valide:
		return "Dossier validé par la compta"
	case Archive:
		return "Dossier archivé"
	default:
		panic("exhaustive switch")
	}
}

// Reviews is a simple conversation between
// admin and director about one record
type Reviews []Review

type Review struct {
	IsAdmin bool // send by the admin ? or by the director ?
	Time    time.Time
	Message string
}

type Cent = int

type Montants []MontantPart

type MontantPart struct {
	Cent Cent
	Kind Ventilation
}

type Ventilation uint8

const (
	NoVentilation Ventilation = iota //

	Activites             // Activités éducatives
	Alimentation          // Alimentation
	FournitureEntretien   // Fournitures d'entretien
	FournitureAdmin       // Fournitures administratives
	FourniturePedagogique // Fournitures pédagogiques
	Medical               // Pharmacie et médecins
	Hotelier              // Prestations hôtelières
	Camping               // Location de camping (ou gîte)
	Carburant             // Carburant
	MaintenanceVehic      // Maintenance des véhicules
	Travels               // Déplacements
	Telephone             // Téléphone
	Poste                 // Frais postaux
	Bank                  // Frais bancaires
	PensionBenevole       // Pension bénévole
	Autre                 // Autre
)

const VentilationCount = Autre + 1 // gomacro:no-enum

type ModePaiementRecord uint8

const (
	CarteB    ModePaiementRecord = iota // Carte bancaire
	Virement                            // Virement
	Liquide                             // Liquide
	Cheque                              // Chèque
	Helloasso                           // Helloasso
)

func (m ModePaiementRecord) String() string {
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

type FileStatus uint8

const (
	FileSystem FileStatus = iota
	GoogleDrive
	Archived
)

type OptIdComptaFile = shared.OptID[IdComptaFile]

func NewOptIdComptaFile(id IdComptaFile) OptIdComptaFile {
	return OptIdComptaFile{Id: id, Valid: true}
}

type OptIdEquipier = shared.OptID[camps.IdEquipier]

type Meal uint8

const (
	PetitDejeuner Meal = iota
	Midi
	Gouter
	Soir
	Cinquieme
)

type NamedCost struct {
	Name string
	Cost dossiers.Montant
}

type Rents []NamedCost

type Activities []NamedCost

// MealsCosts stores the cost, in € cent , of one meal
type MealsCosts [Cinquieme + 1]int32

type CarKind uint8

const (
	CarBus     CarKind = iota // Bus
	CarMinibus                // Minibus
	CarCar                    // Véhicule léger
)

const nbCarKind = CarCar + 1

// CarsCosts stores the kilometric cost for each kind of cars,
// as € cents
type CarsCosts [nbCarKind]int32

type CarsKms [nbCarKind]int32

func (kms CarsKms) Resolve(costs CarsCosts) ds.Montant {
	return ds.Montant{
		Cent:     Cent(kms[CarBus]*costs[CarBus] + kms[CarMinibus]*costs[CarMinibus] + kms[CarCar]*costs[CarCar]),
		Currency: ds.Euros,
	}
}
