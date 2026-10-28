package compta

import (
	"time"

	"registro/sql/dossiers"
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

type Montants []MontantPart

type MontantPart struct {
	dossiers.Montant
	V Ventilation
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

var ventilationLabels = [Autre + 1]struct{ code, label string }{
	NoVentilation: {"", ""},

	Activites:             {"6048", "Activités éducatives"},
	Alimentation:          {"6062", "Alimentation"},
	FournitureEntretien:   {"6063", "Fournitures d'entretien"},
	FournitureAdmin:       {"6064", "Fournitures administratives"},
	FourniturePedagogique: {"60685", "Fournitures éducatives"},
	Medical:               {"60681", "Pharmacie et médecins"},
	Hotelier:              {"6115", "Prestations hôtelières"},
	Camping:               {"613", "Location de camping (ou gîte)"},
	Carburant:             {"6066", "Carburant"},
	MaintenanceVehic:      {"6156", "Maintenance des véhicules"},
	Travels:               {"6251", "Déplacements"},
	Telephone:             {"6262", "Téléphone"},
	Poste:                 {"6263", "Frais postaux"},
	Bank:                  {"627", "Frais bancaires"},
	PensionBenevole:       {"7062", "Pension bénévole"},
	Autre:                 {"AUTRE", "Autre"},
}

func (v Ventilation) forCars() bool {
	return v == Carburant
}

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
