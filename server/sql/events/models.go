package events

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	"registro/sql/camps"
	"registro/sql/dossiers"
)

type IdEvent int64

// Event encode un échange entre le centre d'inscription
// et le responsable d'un dossier
//
// Requis pour référence
// gomacro:SQL ADD UNIQUE(Id, Kind)
type Event struct {
	Id        IdEvent
	IdDossier dossiers.IdDossier `gomacro-sql-on-delete:"CASCADE"`
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

	Contenu     string
	Origine     MessageOrigine
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

	IdCamp camps.IdCamp `gomacro-sql-on-delete:"CASCADE"`
}

// EventCampDocs indique le camp concerné par l'envoi des documents.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.CampDocs])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type EventCampDocs struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  camps.IdCamp

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
	IdCamp  camps.IdCamp

	// For consistency
	Guard EventKind
}

// EventPlaceLiberee notifie qu'un participant a une place disponible.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// contraintes d'intégrité :
// gomacro:SQL ADD CHECK(Guard = #[EventKind.PlaceLiberee])
// gomacro:SQL ADD FOREIGN KEY (IdEvent, Guard) REFERENCES Event(Id,Kind)
type EventPlaceLiberee struct {
	IdEvent       IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdParticipant camps.IdParticipant

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
	IdEvent      IdEvent `gomacro-sql-on-delete:"CASCADE"`
	Distribution Distribution

	// For consistency
	Guard EventKind
}
