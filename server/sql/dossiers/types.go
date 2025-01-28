package dossiers

import (
	"registro/sql/camps"
	"registro/sql/personnes"
)

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

// Mode de paiement
type ModePaiement uint8

const (
	Cheque  ModePaiement = iota
	EnLigne              // (carte bancaire, en ligne)
	Virement
	Especes
	Ancv
	// uniquement pour les dons
	Helloasso
)

type ParticipantExt struct {
	Participant
	camps.Camp
	personnes.Personne
}

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
	ReducSpeciale  camps.Montant
}
