package events

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	cps "registro/sql/camps"
	"registro/sql/dossiers"
)

type IdEvent int64

// Event encode un échange entre le centre d'inscription
// et le responsable d'un dossier
//
// Requis pour référence
// gomacro:SQL ADD UNIQUE(Id, Kind)
//
// gomacro:QUERY SwitchValidationAndMessageDossier UPDATE Event SET IdDossier = $to$ WHERE IdDossier = $from$ AND (Kind = #[EventKind.Message] OR Kind = #[EventKind.Statuation]);
type Event struct {
	Id        IdEvent
	IdDossier dossiers.IdDossier `gomacro-sql-on-delete:"CASCADE"`
	Kind      EventKind
	Created   time.Time
}

// EventStatuation indicates the origin and camp of the validation
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventStatuation struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`

	// le camp du participant (au moment de la validation)
	IdCamp cps.IdCamp
	// true si la validation a été effectuée par le centre
	IsBackoffice bool

	IdParticipant cps.IdParticipant     `gomacro-sql-on-delete:"CASCADE"`
	Statut        cps.StatutParticipant // the decision at statuation time

	guard EventKind `gomacro-sql-guard:"#[EventKind.Statuation]"`
}

// EventMessage stocke le contenu d'un message libre
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
//
// gomacro:SQL ADD CHECK(Origine <> #[Acteur.Directeur] OR OrigineCamp IS NOT NULL)
// gomacro:SQL ADD CHECK(Origine = #[Acteur.Directeur] OR OrigineCamp IS NULL)
//
// gomacro:SQL ADD CHECK(OnlyToFondSoutien = False OR OnlyToCamp IS NULL)
// gomacro:SQL ADD CHECK((OnlyToFondSoutien = False AND OnlyToCamp IS NULL) OR Origine = #[Acteur.Espaceperso])
type EventMessage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`

	Contenu     string
	Origine     Acteur
	OrigineCamp OptIdCamp `gomacro-sql-foreign:"Camp"`

	VuBackoffice  bool
	VuEspaceperso bool
	VuFondSoutien bool

	// OnlyToFondSoutien est utilisé pour restreindre la visibilité d'un message
	// au fonds de soutien.
	// Ce champ n'est utilisé que pour les messages avec Origine == Espaceperso
	OnlyToFondSoutien bool
	// OnlyToCamp est utilisé pour restreindre la visibilité d'un message
	// à un seul directeur (et au centre)
	// Ce champ n'est utilisé que pour les messages avec Origine == Espaceperso
	OnlyToCamp OptIdCamp `gomacro-sql-foreign:"Camp"`

	guard EventKind `gomacro-sql-guard:"#[EventKind.Message]"`
}

// CreateMessage does not wrap errors
func CreateMessage(db DB, idDossier dossiers.IdDossier, created time.Time, message EventMessage) (Event, EventMessage, error) {
	event, err := Event{IdDossier: idDossier, Kind: Message, Created: created}.Insert(db)
	if err != nil {
		return Event{}, EventMessage{}, err
	}
	message.IdEvent = event.Id
	err = message.Insert(db)
	return event, message, err
}

// EventMessageView indique qu'un message a été lu par le directeur.
//
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
// gomacro:SQL ADD UNIQUE(IdEvent, IdCamp)
type EventMessageVu struct {
	IdEvent IdEvent    `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  cps.IdCamp `gomacro-sql-on-delete:"CASCADE"`

	guard EventKind `gomacro-sql-guard:"#[EventKind.Message]"`
}

// EventCampDocs indique le séjour concerné par l'envoi des documents.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventCampDocs struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  cps.IdCamp

	guard EventKind `gomacro-sql-guard:"#[EventKind.CampDocs]"`
}

// EventSondage indique le séjour concerné par le sondage.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventSondage struct {
	IdEvent IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdCamp  cps.IdCamp

	guard EventKind `gomacro-sql-guard:"#[EventKind.Sondage]"`
}

// EventPlaceLiberee notifie qu'un participant a une place disponible.
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventPlaceLiberee struct {
	IdEvent       IdEvent `gomacro-sql-on-delete:"CASCADE"`
	IdParticipant cps.IdParticipant
	Accepted      bool

	guard EventKind `gomacro-sql-guard:"#[EventKind.PlaceLiberee]"`
}

// EventAttestation complète l'accès
// à une facture acquittée/attestation de présence
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventAttestation struct {
	IdEvent      IdEvent `gomacro-sql-on-delete:"CASCADE"`
	Distribution Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool

	guard EventKind `gomacro-sql-guard:"#[EventKind.Attestation]"`
}

// EventChangementCamp rappelle quand un participant a changé de camp
//
// gomacro:SQL ADD UNIQUE(IdEvent)
// gomacro:SQL ADD FOREIGN KEY (IdEvent, guard) REFERENCES Event(Id,Kind) ON DELETE CASCADE
type EventChangementCamp struct {
	IdEvent       IdEvent           `gomacro-sql-on-delete:"CASCADE"`
	IdParticipant cps.IdParticipant `gomacro-sql-on-delete:"CASCADE"`
	Old, New      cps.IdCamp

	guard EventKind `gomacro-sql-guard:"#[EventKind.ChangementCamp]"`
}
