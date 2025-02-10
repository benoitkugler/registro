package events 

type IdEvent int64


// Event encode un échange entre le centre d'inscription
// et le responsable d'un dossier
//
// Requis pour référence
// gomacro:SQL ADD UNIQUE(Id, Kind)
type Event struct {
	Id        IdEvent
	IdDossier ds.IdDossier `gomacro-sql-on-delete:"CASCADE"`
	Kind      EventKind
	Created   time.Time
}

// EventMessage stocke le contenu d'un message libre
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.Message])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
//
// gomacro:SQL ADD CHECK(Origine <> #[MessageOrigine.FromDirecteur] OR OrigineCamp IS NOT NULL)
// gomacro:SQL ADD CHECK(Origine = #[MessageOrigine.FromDirecteur] OR OrigineCamp IS NULL)
type EventMessage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	// For consistency
	Guard EventKind

	Contenu string
	Origine MessageOrigine
	OrigineCamp OptIdCamp

	VuBackoffice  bool 
	VuEspaceperso bool
}

// EventMessageView indique qu'un message a été lue par le directeur.
//
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.Message])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
// sql:ADD UNIQUE(IdEvent, IdCamp)
type EventMessageVu struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	// For consistency
	Guard EventKind

	IdCamp  cps.IdCamp  `gomacro-sql-on-delete:"CASCADE"`
}

// EventCampDocs indique le camp concerné par l'envoi des documents.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.CampDocs])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type EventCampDocs struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  cps.IdCamp

	// For consistency
	Guard EventKind
}

// EventSondage indique le camp concerné par le sondage.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.Sondage])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type EventSondage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  cps.IdCamp

	// For consistency
	Guard EventKind
}

// EventPlacelibere notifie qu'un participant a une place disponible.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.Placelibere])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type EventPlacelibere struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdParticipant cps.IdParticipant 

	// For consistency
	Guard EventKind
}

// MessageAttestation complète l'accès
// à une facture acquittée/attestation de présence
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guard_kind = #MessageKind.MFactureAcquittee OR guard_kind = #MessageKind.MAttestationPresence)
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.Sondage])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type MessageAttestation struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	Distribution Distribution `json:"distribution"`

	GuardKind MessageKind `json:"guard_kind"`
}