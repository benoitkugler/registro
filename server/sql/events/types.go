package events 

type EventKind uint8

const (
	Supprime EventKind = iota // Message supprimé

	// peut provenir du backoffice, du portail directeurs
	// ou de l'espace perso
	Message // Message

	// enregistre le moment d'inscription
	Inscription         // Moment d'inscription
	AccuseReception     // Inscription validée
	Facture             // Facture
	CampDocs           // Document des camps
	FactureAcquittee    // Facture acquittée
	AttestationPresence // Attestation de présence
	Sondage             // Avis sur le séjour

	PlaceLiberee // Place libérée
)

type MessageOrigine uint8

const (
	FromEspaceperso MessageOrigine = iota
	FromBackoffice
	FromDirecteur
)

const (
	DEspacePerso     Distribution = iota // Téléchargée depuis l'espace de suivi
	DMail                                // Notifiée par courriel
	DMailAndDownload                     // Téléchargée après notification
)