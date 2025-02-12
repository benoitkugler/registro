package events

import (
	"registro/sql/camps"
	"registro/sql/shared"
)

type OptIdCamp shared.OptID[camps.IdCamp]

type EventKind uint8

const (
	Supprime EventKind = iota // Message supprimé

	// enregistre le moment d'inscription
	Inscription     // Moment d'inscription
	AccuseReception // Inscription validée

	// peut provenir du backoffice, du portail directeurs
	// ou de l'espace perso
	Message      // Message
	PlaceLiberee // Place libérée
	Facture      // Facture
	CampDocs     // Document des camps
	Attestation  // Facture acquittée ou attestation de présence
	Sondage      // Avis sur le séjour

)

type MessageOrigine uint8

const (
	FromEspaceperso MessageOrigine = iota
	FromBackoffice
	FromDirecteur
)

type Distribution uint8

const (
	DEspacePerso     Distribution = iota // Téléchargée depuis l'espace de suivi
	DMail                                // Notifiée par courriel
	DMailAndDownload                     // Téléchargée après notification
)
