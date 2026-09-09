package personnes

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"registro/sql/shared"
)

type OptIdPersonne = shared.OptID[IdPersonne]

func (id IdPersonne) Opt() OptIdPersonne { return OptIdPersonne{Id: id, Valid: true} }

// Time is date and time
type Time time.Time

// Pays is the ISO 3166 (2 letter) code of a country
type Pays string

// Departement is the number of a french departement,
// or its name for other countries
type Departement string

// Sexe is Man, Woman or undefined.
type Sexe uint8

const (
	NoSexe Sexe = iota //
	Woman              // Femme
	Man                // Homme
)

// Tel is a phone number
type Tel string

// Tels is a list of phone numbers
type Tels [2]string

// Identite stores information about the identity of one person.
//
// The fields defined here are the ones used in profil merging for
// new inscriptions and donateurs
type Identite struct {
	Nom    string
	Prenom string
	Sexe   Sexe

	DateNaissance shared.Date

	Nationnalite Nationnalite

	Tels Tels
	Mail string

	Adresse    string
	CodePostal string
	Ville      string
	Pays       Pays
}

func (p *Identite) FPrenom() string { return formatPrenom(p.Prenom) }

func (p *Identite) FNom() string { return strings.ToUpper(p.Nom) }

// NOMPrenom return NOM Prenom
func (p Identite) NOMPrenom() string {
	return p.FNom() + " " + p.FPrenom()
}

// PrenomNOM return NOM Prenom
func (p Identite) PrenomNOM() string {
	return p.FPrenom() + "  " + p.FNom()
}

// PrenomN returns Prenom N.
func (p *Identite) PrenomN() string {
	var initiale string
	if nom := p.FNom(); nom != "" {
		r, _ := utf8.DecodeRuneInString(nom)
		initiale = string(r)
	}
	return fmt.Sprintf("%s %s.", p.FPrenom(), initiale)
}

// Nationnalite encode la nationnalité,
// qui peut être différente du [Pays]
type Nationnalite struct {
	IsSuisse bool
}

// Formation professionnelle jugée équivalente
type FormationRepere uint8

const (
	FRAucun                      FormationRepere = iota //
	FRCfcAssistantESocioEducatif                        // CFC assistant(e) socio-éducatif
	FRMaitreSocioProfessionnel                          // Maitre socio-professionnel
	FREnseignementScolaire                              // Enseignement scolaire
	FRTravailSocial                                     // Travail social
)

type Diplome uint8

const (
	DAucun          Diplome = iota // Aucun
	DAcveBafa                      // BAFA Titulaire
	DAcveBafaStag                  // BAFA Stagiaire
	DAcveBafd                      // BAFD titulaire
	DAcveBafdStag                  // BAFD stagiaire
	DAcveCap                       // CAP petit enfance
	DAcveAssSociale                // Assitante Sociale
	DAcveEducSpe                   // Educ. spé.
	DAcveMonEduc                   // Moniteur educateur
	DAcveInstit                    // Professeur des écoles
	DAcveProf                      // Enseignant du secondaire
	DAcveAgreg                     // Agrégé
	DAcveBjeps                     // BPJEPS
	DAcveDut                       // DUT carrière sociale
	DAcveEje                       // EJE
	DAcveDeug                      // DEUG
	DAcveStaps                     // STAPS
	DAcveBapaat                    // BAPAAT
	DAcveBeatep                    // BEATEP

	DRepereJsMoniteur  // Formation JS Moniteur
	DRepereJsDirecteur // Formation JS Directeur
	DRepereBafa        // BAFA
	DRepereBafd        // BAFD
	DRepereForje       // Formation Forje (GLAJ)

	DAutre // Autre (à préciser)
)

// TODO: probably remove (see https://github.com/benoitkugler/registro/issues/6)
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

type EtatCivil uint8

const (
	NoEtatCivil EtatCivil = iota //
	Marie                        // Marié(e)
	Celibataire                  // Célibataire
)

type Recommandation struct {
	Nom, Prenom, Mail string
	Tel               Tel
}

//--------------------------------------------------------------------
//------------------------ Fiche Sanitaire ---------------------------
//--------------------------------------------------------------------

type Maladies struct {
	Rubeole    bool
	Varicelle  bool
	Angine     bool
	Oreillons  bool
	Scarlatine bool
	Coqueluche bool
	Otite      bool
	Rougeole   bool
	Rhumatisme bool
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
	Asthme          bool
	Alimentaires    bool
	Medicamenteuses bool
	Autres          string
	ConduiteATenir  string
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

type NomTel struct {
	Nom string
	Tel Tel
}
