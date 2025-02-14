package personnes

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"registro/sql/shared"
)

// Time is date and time
type Time time.Time

// Pays is the ISO 3166 code of a country
type Pays string

// Departement is the number of a french departement,
// or its name for other countries
type Departement string

// Sexe is Man, Woman or undefined.
type Sexe uint8

const (
	Empty Sexe = iota //
	Woman 			  // Femme
	Man        		  // Homme
)

// Tel is a phone number
type Tel string

// Tels is a list of phone numbers
type Tels []string

// Etatcivil stores information about the identity of one person.
//
// The fields defined here are the ones used in profil merging.
type Etatcivil struct {
	Nom    string
	Prenom string
	Sexe   Sexe

	DateNaissance        shared.Date
	VilleNaissance       string
	DepartementNaissance Departement
	Nationnalite         Nationnalite

	Tels Tels
	Mail string

	Adresse    string
	CodePostal string
	Ville      string
	Pays       Pays

	SecuriteSociale string

	NomJeuneFille     string            // used for equipiers
	Profession        string            // used for equipiers
	Etudiant          bool              // used for equipiers
	Fonctionnaire     bool              // used for equipiers
	Diplome           Diplome           // used for equipiers
	Approfondissement Approfondissement // used for equipiers
}

func (p *Etatcivil) FPrenom() string { return formatPrenom(p.Prenom) }

func (p *Etatcivil) FNom() string { return strings.ToUpper(p.Nom) }

// NomPrenom return NOM Prenom
func (p *Etatcivil) NomPrenom() string {
	return p.FNom() + " " + p.FPrenom()
}

// PrenomN returns Prenom N.
func (p *Etatcivil) PrenomN() string {
	var initiale string
	if nom := p.FNom(); nom != "" {
		r, _ := utf8.DecodeRuneInString(initiale)
		initiale = string(r)
	}
	return fmt.Sprintf("%s %s.", p.FPrenom(), initiale)
}

// Nationnalite encode la nationnalité,
// qui peut être différente du [Pays]
type Nationnalite uint8

const (
	Autre Nationnalite = iota
	Francaise
	Suisse
)

type Diplome uint8

const (
	DAucun      Diplome = iota // Aucun
	DBafa                      // BAFA Titulaire
	DBafaStag                  // BAFA Stagiaire
	DBafd                      // BAFD titulaire
	DBafdStag                  // BAFD stagiaire
	DCap                       // CAP petit enfance
	DAssSociale                // Assitante Sociale
	DEducSpe                   // Educ. spé.
	DMonEduc                   // Moniteur educateur
	DInstit                    // Professeur des écoles
	DProf                      // Enseignant du secondaire
	DAgreg                     // Agrégé
	DBjeps                     // BPJEPS
	DDut                       // DUT carrière sociale
	DEje                       // EJE
	DDeug                      // DEUG
	DStaps                     // STAPS
	DBapaat                    // BAPAAT
	DBeatep                    // BEATEP
	DZzautre                   // AUTRE
)

type Approfondissement uint8

const (
	AAucun Approfondissement = iota // Non effectué
	AAutre                          // Approfondissement
	ASb                             // Surveillant de baignade
	ACanoe                          // Canoë - Kayak
	AVoile                          // Voile
	AMoto                           // Loisirs motocyclistes
)

type Mails []string

// Publicite indique les préférences de communication
type Publicite struct {
	VersionPapier bool
	PubHiver      bool
	PubEte        bool
	EchoRocher    bool
	Eonews        bool
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
