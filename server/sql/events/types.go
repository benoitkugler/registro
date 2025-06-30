package events

import (
	"registro/sql/camps"
)

type OptIdCamp = camps.OptIdCamp

type EventKind uint8

const (
	Supprime EventKind = iota // Message supprimé

	Validation // Inscription validée

	// peut provenir du backoffice, du portail directeurs
	// ou de l'espace perso
	Message      // Message
	PlaceLiberee // Place libérée
	Facture      // Facture
	CampDocs     // Document des camps
	Attestation  // Facture acquittée ou attestation de présence
	Sondage      // Avis sur le séjour

)

type Acteur uint8

const (
	Espaceperso Acteur = iota
	Backoffice
	Fondsoutien
	Directeur
)

type Distribution uint8

const (
	DEspacePerso       Distribution = iota // Téléchargée depuis l'espace de suivi
	DMail                                  // Notifiée par courriel
	DMailAndDownloaded                     // Téléchargée après notification
)
