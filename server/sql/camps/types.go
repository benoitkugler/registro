package camps

import (
	"fmt"
	"sort"
	"strings"

	"registro/sql/dossiers"
	"registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
)

type (
	Montant   = dossiers.Montant
	OptIdTaux = dossiers.OptIdTaux
)

// ListeAttente définit le statut d'un participant
// par rapport à la liste d'attente
type ListeAttente uint8

const (
	// personne n'a encore décidé ou placer le participant
	ADecider ListeAttente = iota
	// le profil ne suit pas les conditions du camp
	AttenteProfilInvalide
	// le camp est déjà complet
	AttenteCampComplet
	// une place s'est libérée et on attend une confirmation
	EnAttenteReponse
	// le participant apparait en liste principale
	Inscrit
)

type Bus uint8

const (
	NoBus Bus = iota
	Aller
	Retour
	AllerRetour
)

func (b Bus) Includes(aller bool) bool {
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

// Semaine précise le choix d'une seule semaine de camp
type Semaine uint8

const (
	Tout     Semaine = iota // Camp complet
	Semaine1                // Semaine 1
	Semaine2                // Semaine 2
)

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
// en évitant un éventuel doublon
func (js Jours) NbJours(datesCamp shared.Plage) int {
	if len(js) == 0 {
		return datesCamp.Duree
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
func (js Jours) CalculePrix(prixParJour []Montant) Montant {
	var total Montant
	for _, i := range js.sorted() {
		if i >= len(prixParJour) { // ne devrait pas arriver
			continue
		}
		total.Add(prixParJour[i])
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
	NoOption OptionPrixKind = iota
	PrixSemaine
	PrixStatut
	PrixJour
)

// OptionPrixCamp stocke une option sur le prix d'un camp. Une seule est effective,
// déterminée par Active
type OptionPrixCamp struct {
	Active OptionPrixKind

	Semaine OptionSemaineCamp
	Statuts []PrixParStatut
	// Prix de chaque jour du camp (souvent constant)
	// Le champ [Prix] du séjour peut être inférieur à la somme
	// pour une remise.
	Jour []Montant
}

type OptionSemaineCamp struct {
	Plage1 shared.Plage
	Plage2 shared.Plage
	Prix1  Montant
	Prix2  Montant
}

type PrixParStatut struct {
	Id          int64
	Prix        Montant
	Statut      string
	Description string
}

// OptionPrixParticipant répond à OptionPrixCamp. L'option est active si :
//   - elle est active dans le camp
//   - elle est non nulle dans le participant
type OptionPrixParticipant struct {
	Semaine  Semaine
	IdStatut int
	Jour     Jours
}

// IsNonZero renvoie `true` si une option est active
// pour la catégorie demandée.
func (op OptionPrixParticipant) IsNonZero(kind OptionPrixKind) bool {
	switch kind {
	case PrixSemaine:
		return op.Semaine != 0
	case PrixStatut:
		return op.IdStatut != 0
	case PrixJour:
		return len(op.Jour) > 0
	default:
		return false
	}
}

// grilleQF définit une grille fixe de prix au quotient familial
// Index 0 : de 1 à 359
// Index 1 : de 360 à 564
// Index 2 : de 565 à 714
// Index 3 : plus de 715
var grilleQF = [...]int{0, 359, 564, 714}

// OptionQuotientFamilial applique un pourcentage sur le prix de base (exprimé en %),
// pour les categories définie par `QuotientFamilial`
//
// Par cohérence avec le prix de base, la dernière valeur vaut toujours 100
// (sauf pour les entrées vides).
type OptionQuotientFamilial [len(grilleQF)]int32

// Percentage renvoie le pourcentage appliqué au prix de base pour le quotient
// familial donné.
func (q OptionQuotientFamilial) Percentage(quotientFamilial int) int {
	N := len(grilleQF)
	for i := 0; i < N-1; i++ {
		q1, q2 := grilleQF[i], grilleQF[i+1]
		if q1 < quotientFamilial && quotientFamilial <= q2 {
			return int(q[i])
		}
	}
	return 100
}

type Navette struct {
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

// InvitationEquipier enregistre si
// l'équipier a validé son profil
type InvitationEquipier uint8

const (
	NonInvite InvitationEquipier = iota
	Invite
	Verifie
)

type OptionnalPlage struct {
	shared.Plage
	Active bool
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

type ParticipantExt struct {
	Participant
	Camp
	personnes.Personne
}
