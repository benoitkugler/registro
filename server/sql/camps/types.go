package camps

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"registro/sql/dossiers"
	"registro/sql/shared"
	"registro/utils"
)

type Montant = dossiers.Montant

type OptIdCamp shared.OptID[IdCamp]

func (id IdCamp) Opt() OptIdCamp { return OptIdCamp{Id: id, Valid: true} }

func (s *OptIdCamp) Scan(src any) error { return (*shared.OptID[IdCamp])(s).Scan(src) }

func (s OptIdCamp) Value() (driver.Value, error) { return (shared.OptID[IdCamp])(s).Value() }

// StatutParticipant définit le statut d'un participant
// par rapport à la liste d'attente
type StatutParticipant uint8

const (
	// personne n'a encore décidé ou placer le participant
	AStatuer StatutParticipant = iota // A statuer
	// définitivment refusé (non concerné par une place libérée)
	Refuse // Refusé définitivement
	// le profil ne suit pas les conditions du camp
	AttenteProfilInvalide // Profil limite
	// le camp est déjà complet
	AttenteCampComplet // Camp complet
	// une place s'est libérée et on attend une confirmation
	EnAttenteReponse // En attente de réponse
	// le participant apparait en liste principale
	Inscrit // Inscrit
)

var (
	_ json.Marshaler   = (StatutParticipant)(0)
	_ json.Unmarshaler = (*StatutParticipant)(nil)
)

// MarshalText makes sure [StatutParticipants] is not encoded as a []byte slice
func (d StatutParticipant) MarshalJSON() ([]byte, error) {
	// By default a slice of StatutParticipant is marshalled as string by Go
	return json.Marshal(uint8(d))
}

func (d *StatutParticipant) UnmarshalJSON(src []byte) error {
	return json.Unmarshal(src, (*uint8)(d))
}

type Navette uint8

const (
	NoBus       Navette = iota // Aucun trajet
	Aller                      // Aller
	Retour                     // Retour
	AllerRetour                // Aller-Retour
)

func (b Navette) Includes(aller bool) bool {
	if aller {
		return b == Aller || b == AllerRetour
	}
	return b == Retour || b == AllerRetour
}

// Remises altère le prix payé par un participant
type Remises struct {
	ReducEquipiers int // en %
	ReducEnfants   int // en %
	ReducSpeciale  Montant
}

// Jours stocke les indexes (0-based) des jours de présence
// d'un participant à un séjour
// Une liste vide indique la présence sur TOUT le séjour.
type Jours []int32

// sorted renvoie une liste triée et unique
func (js Jours) sorted() []int {
	set := map[int]bool{}
	for _, index := range js {
		set[int(index)] = true
	}
	sortedKeys := utils.MapKeys(set)
	sort.Ints(sortedKeys)
	return sortedKeys
}

// Sanitize vérifie que les jours sont valides
func (js Jours) Sanitize(duree int) error {
	if len(js.sorted()) != len(js) {
		return fmt.Errorf("jours de présences invalides: %v", js)
	}
	for _, j := range js {
		if j < 0 || int(j) >= duree {
			return fmt.Errorf("jour de présence invalide (%d)", j)
		}
	}
	return nil
}

// NbJours renvoie le nombre de jours de présence
func (js Jours) NbJours(campDuree int) int {
	if len(js) == 0 {
		return campDuree
	}
	return len(js.sorted())
}

// ClosestPlage renvoie la plage englobant les jours de présence
func (js Jours) ClosestPlage(datesCamp shared.Plage) shared.Plage {
	sorted := js.sorted()
	if len(sorted) == 0 { // zero value : tout le séjour
		return datesCamp
	}
	indexMin, indexMax := sorted[0], sorted[len(js)-1]
	return shared.Plage{From: datesCamp.From.AddDays(indexMin), Duree: indexMax - indexMin + 1}
}

// CalculePrix somme les prix des journées de présence
func (js Jours) CalculePrix(prixParJour []int, currency dossiers.Currency) Montant {
	total := Montant{Currency: currency}
	for _, i := range js.sorted() {
		// npeut arriver si la durée d'un séjour est réduite
		// après avoir déclaré une option sur un participant
		if i >= len(prixParJour) {
			continue
		}
		total.Cent += prixParJour[i]
	}
	return total
}

// Description renvoie les jours de présence au camp
func (js Jours) Description(datesCamp shared.Plage) string {
	sorted := js.sorted()
	if len(sorted) == 0 || len(sorted) == datesCamp.Duree {
		return "Tout le séjour"
	}

	var days []string
	for _, index := range sorted { // 0 based
		day := datesCamp.From.AddDays(index)
		days = append(days, day.ShortString())
	}
	return strings.Join(days, "; ")
}

type OptionPrixKind uint8

const (
	NoOption   OptionPrixKind = iota // Aucune
	PrixStatut                       // Prix par statut
	PrixJour                         // Prix à la journée
)

// OptionPrixCamp stocke une option sur le prix d'un camp. Une seule est effective,
// déterminée par Active
type OptionPrixCamp struct {
	Active OptionPrixKind

	Statuts []PrixParStatut

	// Prix de chaque jour (0-based) du camp (souvent constant), en centimes.
	// L'unité est celle du séjour associé.
	// Le champ [Prix] du séjour peut être inférieur à la somme
	// pour une remise.
	Jours []int
}

type PrixParStatut struct {
	Id          int16
	Prix        int // prix en centimes (l'unité est celle du camp)
	Label       string
	Description string // longue description
}

// OptionPrixParticipant répond à OptionPrixCamp. L'option est active si :
//   - elle est active dans le camp
//   - elle est non nulle dans le participant
type OptionPrixParticipant struct {
	IdStatut int16
	Jour     Jours
}

// IsEmpty renvoie `true` si aucune option n'est active
// pour la catégorie demandée.
func (op OptionPrixParticipant) IsEmpty(kind OptionPrixKind) bool {
	switch kind {
	case PrixStatut:
		return op.IdStatut == 0
	case PrixJour:
		return len(op.Jour) == 0
	default:
		return true
	}
}

// grilleQF définit une grille fixe de prix au quotient familial
// Index 0 : de 1 à 359
// Index 1 : de 360 à 564
// Index 2 : de 565 à 714
// Index 3 : plus de 715
var grilleQF = [...]int{0, 359, 564, 714}

// PrixQuotientFamilial applique un pourcentage sur le prix de base (exprimé en %),
// pour les categories définie par `QuotientFamilial`
//
// Par cohérence avec le prix de base, la dernière valeur vaut toujours 100
// (sauf pour la valeur zero).
type PrixQuotientFamilial [len(grilleQF)]int32

// IsActive renvoie 'true' si la réduction est active
func (oq PrixQuotientFamilial) IsActive() bool { return oq[3] == 100 }

// Percentage renvoie le pourcentage appliqué au prix de base pour le quotient
// familial donné.
func (q PrixQuotientFamilial) Percentage(quotientFamilial int) int {
	N := len(grilleQF)
	for i := 0; i < N-1; i++ {
		q1, q2 := grilleQF[i], grilleQF[i+1]
		if q1 < quotientFamilial && quotientFamilial <= q2 {
			return int(q[i])
		}
	}
	return 100
}

type OptionNavette struct {
	Actif       bool
	Commentaire string
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
	Direction                 // Direction
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

const NbRoles = Lingerie + 1 // gomacro:no-enum

func (r Role) String() string {
	switch r {
	case AutreRole:
		return "Autre"
	case Direction:
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
	case Direction, Adjoint, Animation, AideAnimation:
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

var (
	_ json.Marshaler   = (Role)(0)
	_ json.Unmarshaler = (*Role)(nil)
)

// MarshalText makes sure [Roles] is not encoded as a []byte slice
func (d Role) MarshalJSON() ([]byte, error) {
	// By default a slice of SignSymbol is marshalled as string by Go
	return json.Marshal(uint8(d))
}

func (d *Role) UnmarshalJSON(src []byte) error {
	return json.Unmarshal(src, (*uint8)(d))
}

// FormStatusEquipier enregistre si
// l'équipier a validé son profil
type FormStatusEquipier uint8

const (
	NotSend  FormStatusEquipier = iota // Non envoyé
	Pending                            // En attente
	Answered                           // Répondu
)

// PresenceOffsets encode une différence par rapport
// à une plage de référence (celle du camp).
//
// La valeur zéro correspond à la date par défaut.
type PresenceOffsets struct {
	// Nombre de jours à ajouter
	Debut, Fin int
}

// Satisfaction est une énumération indiquant le
// niveau de satisfaction sur le sondage de fin de séjour
type Satisfaction uint8

// Attention, la valeur compte pour la présentation
// sur le frontend comme "form-rating"
const (
	NoSatisfaction   Satisfaction = iota // -
	Decevant                             // Décevant
	Moyen                                // Moyen
	Satisfaisant                         // Satisfaisant
	Tressatisfaisant                     // Très satisfaisant
)

type ReponseSondage struct {
	InfosAvantSejour   Satisfaction
	InfosPendantSejour Satisfaction
	Hebergement        Satisfaction
	Activites          Satisfaction
	Theme              Satisfaction
	Nourriture         Satisfaction
	Hygiene            Satisfaction
	Ambiance           Satisfaction
	Ressenti           Satisfaction
	MessageEnfant      string
	MessageResponsable string
}
