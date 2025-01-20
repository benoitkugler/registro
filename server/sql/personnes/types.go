package personnes

import (
	"strings"
	"time"
)

// Time is date and time
type Time time.Time

// Date is a date (without notion of time)
type Date time.Time

// Country is the ISO 3166 code of a country
type Country string

// Departement is the number of a french departement,
// or its name for other countries
type Departement string

// Sex is Man, Woman or undefined.
type Sex uint8

const (
	_ Sex = iota
	Woman
	Man
)

// Tel is a phone number
type Tel string

// Tels is a list of phone numbers
type Tels []string

// Etatcivil stores information about the identity of one person.
type Etatcivil struct {
	Nom                  string
	NomJeuneFille        string
	Prenom               string
	DateNaissance        Date
	VilleNaissance       string
	DepartementNaissance Departement
	Sexe                 Sex
	Tels                 Tels
	Mail                 string
	Adresse              string
	CodePostal           string
	Ville                string
	Pays                 Country
	SecuriteSociale      string
	Profession           string
	Etudiant             bool
	Fonctionnaire        bool
}

func (p Etatcivil) FPrenom() string { return formatPrenom(p.Prenom) }

func (p Etatcivil) FNom() string { return strings.ToUpper(p.Nom) }

// NomPrenom return NOM Prenom
func (p Etatcivil) NomPrenom() string {
	return p.FNom() + " " + p.FPrenom()
}

//--------------------------------------------------------------------
//------------------------ Fiche Sanitaire ---------------------------
//--------------------------------------------------------------------

type Maladies struct {
	Rubeole    bool `json:"rubeole"`
	Varicelle  bool `json:"varicelle"`
	Angine     bool `json:"angine"`
	Oreillons  bool `json:"oreillons"`
	Scarlatine bool `json:"scarlatine"`
	Coqueluche bool `json:"coqueluche"`
	Otite      bool `json:"otite"`
	Rougeole   bool `json:"rougeole"`
	Rhumatisme bool `json:"rhumatisme"`
}

// List returns the checked diseases, as a list.
func (m Maladies) List() []string {
	var out []string
	if m.Rubeole {
		out = append(out, "Rubéole")
	}
	if m.Varicelle {
		out = append(out, "Varicelle")
	}
	if m.Angine {
		out = append(out, "Angine")
	}
	if m.Oreillons {
		out = append(out, "Oreillons")
	}
	if m.Scarlatine {
		out = append(out, "Scarlatine")
	}
	if m.Coqueluche {
		out = append(out, "Coqueluche")
	}
	if m.Otite {
		out = append(out, "Otite")
	}
	if m.Rougeole {
		out = append(out, "Rougeole")
	}
	if m.Rhumatisme {
		out = append(out, "Rhumatisme articulaire aigü")
	}
	return out
}

type Allergies struct {
	Asthme          bool   `json:"asthme"`
	Alimentaires    bool   `json:"alimentaires"`
	Medicamenteuses bool   `json:"medicamenteuses"`
	Autres          string `json:"autres"`
	ConduiteATenir  string `json:"conduite_a_tenir"`
}

func (a Allergies) List() []string {
	var out []string
	if a.Asthme {
		out = append(out, "Asthme")
	}
	if a.Alimentaires {
		out = append(out, "Alimentaires")
	}
	if a.Medicamenteuses {
		out = append(out, "Médicamenteuses")
	}
	if a.Autres != "" {
		out = append(out, a.Autres)
	}
	return out
}

type Medecin struct {
	Nom string `json:"nom"`
	Tel Tel    `json:"tel"`
}

// FicheSanitaire stores information as declared on the personnal space.
// Information from the responsable legal will be required to display
// the complete document.
type FicheSanitaire struct {
	TraitementMedical bool      `json:"traitement_medical"`
	Maladies          Maladies  `json:"maladies"`
	Allergies         Allergies `json:"allergies"`
	DifficultesSante  string    `json:"difficultes_sante"`
	Recommandations   string    `json:"recommandations"`
	Handicap          bool      `json:"handicap"`
	Tel               Tel       `json:"tel"` // added to the one of the responsable
	Medecin           Medecin   `json:"medecin"`

	LastModif Time     `json:"last_modif"` // dernière modification
	Mails     []string `json:"mails"`      // owners
}
