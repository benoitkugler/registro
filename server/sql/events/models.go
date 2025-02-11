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
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
//
// gomacro:SQL ADD CHECK(Origine <> #[MessageOrigine.FromDirecteur] OR OrigineCamp IS NOT NULL)
// gomacro:SQL ADD CHECK(Origine = #[MessageOrigine.FromDirecteur] OR OrigineCamp IS NULL)
type EventMessage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`

	Contenu     string
	Origine     MessageOrigine
	OrigineCamp OptIdCamp

	VuBackoffice  bool
	VuEspaceperso bool

	guard EventKind `gomacro-sql-guard:"#[EventKind.Message]"`
}

// EventMessageView indique qu'un message a été lue par le directeur.
//
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
// gomacro:SQL ADD UNIQUE(IdEvent, IdCamp)
type EventMessageVu struct {
	IdEvent IdEvent      `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  camps.IdCamp `gomacro-sql-on-delete:"CASCADE"`

	guard EventKind `gomacro-sql-guard:"#[EventKind.Message]"`
}

// EventCampDocs indique le camp concerné par l'envoi des documents.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
type EventCampDocs struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  camps.IdCamp

	guard EventKind `gomacro-sql-guard:"#[EventKind.CampDocs]"`
}

// EventSondage indique le camp concerné par le sondage.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
type EventSondage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  camps.IdCamp

	guard EventKind `gomacro-sql-guard:"#[EventKind.Sondage]"`
}

// EventPlaceLiberee notifie qu'un participant a une place disponible.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
type EventPlaceLiberee struct {
	IdEvent       IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdParticipant camps.IdParticipant

	guard EventKind `gomacro-sql-guard:"#[EventKind.PlaceLiberee]"`
}

// EventAttestation complète l'accès
// à une facture acquittée/attestation de présence
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind)
type EventAttestation struct {
	IdEvent      IdEvent `gomacro-sql-on-delete:"CASCADE"`
	Distribution Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool

	guard EventKind `gomacro-sql-guard:"#[EventKind.Attestation]"`
}
