package compta

import (
	"errors"
	"time"

	"github.com/benoitkugler/goACVE/server/shared"
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

func computeState(compta ComptaCamp, records Records) State {
	if compta.State == ToBeValidated {
		// are all records valid ?
		if records.allValidated() {
			return Valide
		}
	}
	if compta.State == EnCours {
		// resolve to Valide if all records exist and are valid
		// even if the director has not ask for
		if len(records) != 0 && records.allValidated() {
			return Valide
		}
	}

	return compta.State
}

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

func (records Records) allValidated() bool {
	for _, rec := range records {
		if !rec.IsValidated {
			return false
		}
	}
	return true
}

// Reviews is a simple conversation between
// admin and director about one record
type Reviews []Review

type Review struct {
	IsAdmin bool // send by the admin ? or by the director ?
	Time    time.Time
	Message string
}

type Montants []Montant

func (ms Montants) equals(other Montants) bool {
	if len(ms) != len(other) {
		return false
	}
	for i := range ms {
		if ms[i] != other[i] {
			return false
		}
	}
	return true
}

func (ms Montants) total() Cent {
	var total Cent
	for _, m := range ms {
		total += m.Montant
	}
	return total
}

func (ms Montants) uniquify() Montants {
	tmp := map[Ventilation]Cent{}
	for _, m := range ms {
		tmp[m.Kind] += m.Montant
	}
	out := make(Montants, 0, len(tmp))
	for k, m := range tmp {
		out = append(out, Montant{m, k})
	}
	return out
}

func (ms Montants) checkVentilations(isReceipe bool) error {
	for _, m := range ms {
		if m.Kind == NoVentilation {
			return shared.FormatErr("Ventilation manquante", errors.New("Une ventilation doit être choisie."))
		}
		if m.Kind == PensionBenevole && !isReceipe {
			return shared.FormatErr("Ventilation incorrecte", errors.New("La ventilation Pension bénévole n'est valide que pour une recette."))
		}
	}
	return nil
}

type Montant struct {
	Montant Cent // in cents
	Kind    Ventilation
}

func (m Montant) toDepense(isReceipe, isOtherAccount bool) depenseTotale {
	if isReceipe {
		if isOtherAccount {
			return depenseTotale{otherAccountRec: m.Montant}
		}
		return depenseTotale{ownAccountRec: m.Montant}
	} else {
		if isOtherAccount {
			return depenseTotale{otherAccount: m.Montant}
		}
		return depenseTotale{ownAccount: m.Montant}
	}
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

func (rec *ComptaRecord) carExpenses() Cent {
	var total Cent
	for _, vent := range rec.Montants {
		if vent.Kind.forCars() {
			total += vent.Montant
		}
	}
	return total
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
