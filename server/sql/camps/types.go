package camps

import (
	"fmt"
	"strings"
)

type Currency uint8

const (
	Empty Currency = iota
	Euros
	FrancsSuisse
)

func (c Currency) String() string {
	switch c {
	case Euros:
		return "€"
	case FrancsSuisse:
		return "CHF"
	default:
		return "<invalid currency>"
	}
}

// Montant représente un prix (avec son unité).
type Montant struct {
	Cent     int
	Currency Currency
}

func NewEuros(f float32) Montant { return Montant{int(f * 100), Euros} }

func (s Montant) String() string {
	return strings.ReplaceAll(fmt.Sprintf("%g %s", float64(s.Cent)/100, s.Currency), ".", ",")
}

type Vetement struct {
	Quantite    int
	Description string
	Important   bool
}

type ListeVetements struct {
	Vetements []Vetement
	// HTML code inserted at the end of the list
	Complement string
}

type Role uint8

const (
	AutreRole     Role = iota // Autre
	Directeur                 // Direction
	Adjoint                   // Adjoint
	Animation                 // Animation
	AideAnimation             // Aide-animateur
	Chauffeur                 // Chauffeur
	Intendance                // Intendance
	Babysiter                 // Baby-sitter
	Menage                    // Ménage
	Factotum                  // Factotum
	Infirmerie                // Assistant sanitaire
	Cuisine                   // Cuisine
	Lingerie                  // Lingerie
)

func (r Role) String() string {
	switch r {
	case AutreRole:
		return "Autre"
	case Directeur:
		return "Direction"
	case Adjoint:
		return "Adjoint"
	case Animation:
		return "Animation"
	case AideAnimation:
		return "Aide-animateur"
	case Chauffeur:
		return "Chauffeur"
	case Intendance:
		return "Intendance"
	case Babysiter:
		return "Baby-sitter"
	case Menage:
		return "Ménage"
	case Factotum:
		return "Factotum"
	case Infirmerie:
		return "Assistant sanitaire"
	case Cuisine:
		return "Cuisine"
	case Lingerie:
		return "Lingerie"
	default:
		return fmt.Sprintf("<role inconnu %d>", r)
	}
}

// IsAuPair renvoie `true` si le rôle est considéré
// comme au pair.
func (r Role) IsAuPair() bool {
	switch r {
	case Directeur, Adjoint, Animation, AideAnimation:
		return true
	default:
		return false
	}
}

type Roles []Role

func (rs Roles) String() string {
	chuncks := make([]string, len(rs))
	for i, v := range rs {
		chuncks[i] = v.String()
	}
	return strings.Join(chuncks, "; ")
}

// IsAuPair vérifie si au moins un des rôles est
// considéré comme au pair
func (rs Roles) IsAuPair() bool {
	for _, r := range rs {
		if r.IsAuPair() {
			return true
		}
	}
	return false
}

// Is vérifie si `r` est présent
func (rs Roles) Is(r Role) bool {
	for _, v := range rs {
		if v == r {
			return true
		}
	}
	return false
}

// InvitationEquipier enregistre si
// l'équipier a validé son profil
type InvitationEquipier uint8

const (
	NonInvite InvitationEquipier = iota
	Invite
	Verifie
)
