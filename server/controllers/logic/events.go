package logic

import (
	"slices"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"
)

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro events.go go/unions:events_gen.go

// Event exposes on event on the dossier track
type Event struct {
	Id      events.IdEvent
	Created time.Time
	Content EventContent
}

// Events stores the [Event] for one [Dossier]
type Events []Event

func (evs Events) By(kind events.EventKind) []Event {
	var out []Event
	for _, ev := range evs {
		if ev.Content.kind() == kind {
			out = append(out, ev)
		}
	}
	return out
}

// NewMessagesForBackoffice returns the [Event]s with kind [events.Message],
// not yet seen by the backoffice
func (evs Events) NewMessagesForBackoffice() (out []Event) {
	for _, ev := range evs {
		if message, ok := ev.Content.(Message); ok {
			if !message.Message.VuBackoffice {
				out = append(out, ev)
			}
		}
	}
	return out
}

// Event exposes on event on the dossier track
type EventContent interface {
	kind() events.EventKind
}

func (Supprime) kind() events.EventKind        { return events.Supprime }
func (AccuseReception) kind() events.EventKind { return events.AccuseReception }
func (Message) kind() events.EventKind         { return events.Message }
func (Facture) kind() events.EventKind         { return events.Facture }
func (CampDocs) kind() events.EventKind        { return events.CampDocs }
func (PlaceLiberee) kind() events.EventKind    { return events.PlaceLiberee }
func (Attestation) kind() events.EventKind     { return events.Attestation }
func (Sondage) kind() events.EventKind         { return events.Sondage }

type (
	Supprime        struct{}
	AccuseReception struct{}
)

type Message struct {
	Message          events.EventMessage
	OrigineCampLabel string   // optionnel
	VuParCamps       []string // labels
}

// m must have kind [Message]
func (ld *EventsData) newMessage(ev events.Event) Message {
	m := ld.messages[ev.Id]
	out := Message{Message: m}
	if m.OrigineCamp.Valid {
		out.OrigineCampLabel = ld.camps[m.OrigineCamp.Id].Label()
	}
	for _, vu := range ld.vupars[m.IdEvent] {
		out.VuParCamps = append(out.VuParCamps, ld.camps[vu.IdCamp].Label())
	}
	return out
}

type Facture struct{}

type CampDocs struct {
	IdCamp    cps.IdCamp
	CampLabel string
}

// m must have kind [CampDocs]
func (ld *EventsData) newCampDocs(ev events.Event) CampDocs {
	m := ld.campDocs[ev.Id]
	camp := ld.camps[m.IdCamp]
	return CampDocs{IdCamp: m.IdCamp, CampLabel: camp.Label()}
}

type PlaceLiberee struct {
	IdParticipant    cps.IdParticipant
	IdCamp           cps.IdCamp
	ParticipantLabel string
	CampLabel        string
}

// m must have kind [PlaceLiberee]
func (ld *EventsData) newPlaceLiberee(ev events.Event) PlaceLiberee {
	m := ld.placeLiberees[ev.Id]
	participant := ld.participants[m.IdParticipant]
	camp := ld.camps[participant.IdCamp]
	pers := ld.personnes[participant.IdPersonne]
	return PlaceLiberee{
		IdParticipant: m.IdParticipant, IdCamp: participant.IdCamp,
		ParticipantLabel: pers.PrenomN(), CampLabel: camp.Label(),
	}
}

type Attestation struct {
	Distribution events.Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool
}

// m must have kind [Attestation]
func (ld *EventsData) newAttestation(ev events.Event) Attestation {
	m := ld.attestations[ev.Id]
	return Attestation{Distribution: m.Distribution, IsPresence: m.IsPresence}
}

type Sondage struct {
	IdCamp    cps.IdCamp
	CampLabel string
}

// m must have kind [Sondage]
func (ld *EventsData) newSondage(ev events.Event) Sondage {
	m := ld.campDocs[ev.Id]
	camp := ld.camps[m.IdCamp]
	return Sondage{IdCamp: m.IdCamp, CampLabel: camp.Label()}
}

type EventsData struct {
	events       map[ds.IdDossier]events.Events
	camps        cps.Camps
	participants cps.Participants
	personnes    pr.Personnes

	messages      map[events.IdEvent]events.EventMessage
	vupars        map[events.IdEvent]events.EventMessageVus
	campDocs      map[events.IdEvent]events.EventCampDocs
	placeLiberees map[events.IdEvent]events.EventPlaceLiberee
	attestations  map[events.IdEvent]events.EventAttestation
	sondages      map[events.IdEvent]events.EventSondage
}

// LoadEvents loads the data required to build the events
// linked to the given dossiers.
//
// It does wrap any error encountered.
func LoadEvents(db events.DB, dossiers ...ds.IdDossier) (out EventsData, _ error) {
	allEvents, err := events.SelectEventsByIdDossiers(db, dossiers...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.events = allEvents.ByIdDossier()
	ids := allEvents.IDs()

	tmp1, err := events.SelectEventMessagesByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.messages = tmp1.ByIdEvent()

	tmp1bis, err := events.SelectEventMessageVusByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.vupars = tmp1bis.ByIdEvent()

	tmp2, err := events.SelectEventCampDocssByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.campDocs = tmp2.ByIdEvent()

	tmp3, err := events.SelectEventPlaceLibereesByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.placeLiberees = tmp3.ByIdEvent()

	tmp4, err := events.SelectEventAttestationsByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.attestations = tmp4.ByIdEvent()

	tmp5, err := events.SelectEventSondagesByIdEvents(db, ids...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.sondages = tmp5.ByIdEvent()

	out.participants, err = cps.SelectParticipants(db, tmp3.IdParticipants()...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}

	var idCamps []cps.IdCamp
	for _, m := range tmp1 {
		if m.OrigineCamp.Valid {
			idCamps = append(idCamps, m.OrigineCamp.Id)
		}
	}
	idCamps = slices.Concat(idCamps, tmp1bis.IdCamps(), tmp5.IdCamps(), out.participants.IdCamps())
	out.camps, err = cps.SelectCamps(db, idCamps...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.personnes, err = pr.SelectPersonnes(db, out.participants.IdPersonnes()...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}

	return out, nil
}

func (ld *EventsData) For(idDossier ds.IdDossier) Events {
	raws := ld.events[idDossier]
	out := make([]Event, 0, len(raws))
	for _, event := range raws {
		val := Event{Id: event.Id, Created: event.Created}
		switch event.Kind {
		case events.Supprime:
			val.Content = Supprime{}
		case events.AccuseReception:
			val.Content = AccuseReception{}
		case events.Message:
			val.Content = ld.newMessage(event)
		case events.PlaceLiberee:
			val.Content = ld.newPlaceLiberee(event)
		case events.Facture:
			val.Content = Facture{}
		case events.CampDocs:
			val.Content = ld.newCampDocs(event)
		case events.Attestation:
			val.Content = ld.newAttestation(event)
		case events.Sondage:
			val.Content = ld.newSondage(event)
		}
		out = append(out, val)
	}
	return out
}
