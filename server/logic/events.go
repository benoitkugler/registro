package logic

import (
	"iter"
	"slices"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"
)

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro events.go go/unions:events_gen.go

// Event exposes on event on the dossier track
type Event struct {
	Id        evs.IdEvent
	idDossier ds.IdDossier
	Created   time.Time
	Content   EventContent
}

func (ev Event) Raw() evs.Event {
	return evs.Event{Id: ev.Id, IdDossier: ev.idDossier, Created: ev.Created, Kind: ev.Content.kind()}
}

// Events stores the [Event] for one [Dossier]
type Events []Event

type EventExt[T EventContent] struct {
	Event   evs.Event
	Content T
}

func IterEventsBy[T EventContent](evs Events) iter.Seq[EventExt[T]] {
	return func(yield func(EventExt[T]) bool) {
		for _, ev := range evs {
			if typed, ok := ev.Content.(T); ok {
				if !yield(EventExt[T]{ev.Raw(), typed}) {
					return
				}
			}
		}
	}
}

// UnreadMessagesForBackoffice returns the [Event]s with kind [evs.Message],
// not yet seen by the backoffice
func (events Events) UnreadMessagesForBackoffice() (out []EventExt[Message]) {
	for ev := range IterEventsBy[Message](events) {
		if ev.Content.Message.Origine != evs.Backoffice && !ev.Content.Message.VuBackoffice {
			out = append(out, ev)
		}
	}
	return out
}

// HasSendCampDocuments returns [true] is the documents for the
// given camp has been sent (at least once)
func (events Events) HasSendCampDocuments(idCamp cps.IdCamp) bool {
	for camp := range IterEventsBy[CampDocs](events) {
		if camp.Content.IdCamp == idCamp {
			return true
		}
	}
	return false
}

// LastBy may return a zero value
func (events Events) LastBy(kind evs.EventKind) Event {
	var last Event
	for _, event := range events {
		if event.Raw().Kind == kind && event.Created.After(last.Created) {
			last = event
		}
	}
	return last
}

// Event exposes on event on the dossier track
type EventContent interface {
	kind() evs.EventKind
}

func (Supprime) kind() evs.EventKind     { return evs.Supprime }
func (Validation) kind() evs.EventKind   { return evs.Validation }
func (Message) kind() evs.EventKind      { return evs.Message }
func (Facture) kind() evs.EventKind      { return evs.Facture }
func (CampDocs) kind() evs.EventKind     { return evs.CampDocs }
func (PlaceLiberee) kind() evs.EventKind { return evs.PlaceLiberee }
func (Attestation) kind() evs.EventKind  { return evs.Attestation }
func (Sondage) kind() evs.EventKind      { return evs.Sondage }

type Supprime struct{}

type Validation struct {
	ByCamp string // optionnel
}

// m must have kind [Validation]
func (ld *eventsContent) newValidation(ev evs.Event) Validation {
	m := ld.validations[ev.Id]
	label := ""
	if m.IdCamp.Valid {
		camp := ld.camps[m.IdCamp.Id]
		label = camp.Label()
	}
	return Validation{label}
}

type Message struct {
	Message          evs.EventMessage
	OrigineCampLabel string // optionnel
	VuParCampsIDs    []cps.IdCamp
	VuParCamps       []string // labels
}

// m must have kind [Message]
func (ld *eventsContent) newMessage(ev evs.Event) Message {
	m := ld.messages[ev.Id]
	out := Message{Message: m}
	if m.OrigineCamp.Valid {
		out.OrigineCampLabel = ld.camps[m.OrigineCamp.Id].Label()
	}
	for _, vu := range ld.vupars[m.IdEvent] {
		out.VuParCampsIDs = append(out.VuParCampsIDs, vu.IdCamp)
		out.VuParCamps = append(out.VuParCamps, ld.camps[vu.IdCamp].Label())
	}
	return out
}

type Facture struct {
	IsRappel bool
}

// check is their is a previous facture sent
func (ld *eventsContent) newFacture(ev evs.Event, all evs.Events) Facture {
	for _, other := range all {
		if other.Id != ev.Id && other.Kind == evs.Facture && other.Created.Before(ev.Created) {
			return Facture{true}
		}
	}
	return Facture{false}
}

type CampDocs struct {
	IdCamp    cps.IdCamp
	CampLabel string
}

// m must have kind [CampDocs]
func (ld *eventsContent) newCampDocs(ev evs.Event) CampDocs {
	m := ld.campDocs[ev.Id]
	camp := ld.camps[m.IdCamp]
	return CampDocs{IdCamp: m.IdCamp, CampLabel: camp.Label()}
}

type PlaceLiberee struct {
	Accepted         bool
	IdParticipant    cps.IdParticipant
	IdCamp           cps.IdCamp
	ParticipantLabel string
	CampLabel        string
}

// m must have kind [PlaceLiberee]
func (ld *eventsContent) newPlaceLiberee(ev evs.Event) PlaceLiberee {
	m := ld.placeLiberees[ev.Id]
	participant := ld.participants[m.IdParticipant]
	camp := ld.camps[participant.IdCamp]
	pers := ld.personnes[participant.IdPersonne]
	return PlaceLiberee{
		m.Accepted,
		m.IdParticipant, participant.IdCamp,
		pers.PrenomN(), camp.Label(),
	}
}

type Attestation struct {
	Distribution evs.Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool
}

// m must have kind [Attestation]
func (ld *eventsContent) newAttestation(ev evs.Event) Attestation {
	m := ld.attestations[ev.Id]
	return Attestation{Distribution: m.Distribution, IsPresence: m.IsPresence}
}

type Sondage struct {
	IdCamp    cps.IdCamp
	CampLabel string
}

// m must have kind [Sondage]
func (ld *eventsContent) newSondage(ev evs.Event) Sondage {
	m := ld.sondages[ev.Id]
	camp := ld.camps[m.IdCamp]
	return Sondage{IdCamp: m.IdCamp, CampLabel: camp.Label()}
}

type eventsContent struct {
	camps        cps.Camps
	participants cps.Participants
	personnes    pr.Personnes

	validations   map[evs.IdEvent]evs.EventValidation
	messages      map[evs.IdEvent]evs.EventMessage
	vupars        map[evs.IdEvent]evs.EventMessageVus
	campDocs      map[evs.IdEvent]evs.EventCampDocs
	placeLiberees map[evs.IdEvent]evs.EventPlaceLiberee
	attestations  map[evs.IdEvent]evs.EventAttestation
	sondages      map[evs.IdEvent]evs.EventSondage
}

// loadEventsContent loads the data required to build the given events.
//
// It does wrap any error encountered.
func loadEventsContent(db evs.DB, ids ...evs.IdEvent) (out eventsContent, _ error) {
	tmp1, err := evs.SelectEventMessagesByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.messages = tmp1.ByIdEvent()

	tmp1bis, err := evs.SelectEventMessageVusByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.vupars = tmp1bis.ByIdEvent()

	tmp20, err := evs.SelectEventValidationsByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.validations = tmp20.ByIdEvent()

	tmp2, err := evs.SelectEventCampDocssByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.campDocs = tmp2.ByIdEvent()

	tmp3, err := evs.SelectEventPlaceLibereesByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.placeLiberees = tmp3.ByIdEvent()

	tmp4, err := evs.SelectEventAttestationsByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.attestations = tmp4.ByIdEvent()

	tmp5, err := evs.SelectEventSondagesByIdEvents(db, ids...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.sondages = tmp5.ByIdEvent()

	out.participants, err = cps.SelectParticipants(db, tmp3.IdParticipants()...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}

	var idCamps []cps.IdCamp
	for _, m := range tmp1 {
		if m.OrigineCamp.Valid {
			idCamps = append(idCamps, m.OrigineCamp.Id)
		}
	}
	for _, m := range tmp20 {
		if m.IdCamp.Valid {
			idCamps = append(idCamps, m.IdCamp.Id)
		}
	}
	idCamps = slices.Concat(idCamps, tmp1bis.IdCamps(), tmp2.IdCamps(), tmp5.IdCamps(), out.participants.IdCamps())
	out.camps, err = cps.SelectCamps(db, idCamps...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}
	out.personnes, err = pr.SelectPersonnes(db, out.participants.IdPersonnes()...)
	if err != nil {
		return eventsContent{}, utils.SQLError(err)
	}

	return out, nil
}

func (ec *eventsContent) build(event evs.Event, dossierEvents evs.Events) Event {
	out := Event{Id: event.Id, idDossier: event.IdDossier, Created: event.Created}
	switch event.Kind {
	case evs.Supprime:
		out.Content = Supprime{}
	case evs.Validation:
		out.Content = ec.newValidation(event)
	case evs.Message:
		out.Content = ec.newMessage(event)
	case evs.PlaceLiberee:
		out.Content = ec.newPlaceLiberee(event)
	case evs.Facture:
		out.Content = ec.newFacture(event, dossierEvents)
	case evs.CampDocs:
		out.Content = ec.newCampDocs(event)
	case evs.Attestation:
		out.Content = ec.newAttestation(event)
	case evs.Sondage:
		out.Content = ec.newSondage(event)
	}
	return out
}

type EventsData struct {
	events map[ds.IdDossier]evs.Events
	eventsContent
}

// LoadEventsByDossier is a convience wrapper which calls
// [LoadEventsByDossiers] for only one dossier.
func LoadEventsByDossier(db evs.DB, dossier ds.IdDossier) (Events, error) {
	loader, err := LoadEventsByDossiers(db, dossier)
	if err != nil {
		return nil, err
	}
	return loader.For(dossier), nil
}

// LoadEventsByDossiers loads the data required to build the events
// linked to the given dossiers.
//
// It does wrap any error encountered.
func LoadEventsByDossiers(db evs.DB, dossiers ...ds.IdDossier) (out EventsData, _ error) {
	allEvents, err := evs.SelectEventsByIdDossiers(db, dossiers...)
	if err != nil {
		return EventsData{}, utils.SQLError(err)
	}
	out.events = allEvents.ByIdDossier()

	out.eventsContent, err = loadEventsContent(db, allEvents.IDs()...)
	if err != nil {
		return EventsData{}, err
	}
	return out, nil
}

func (ld *EventsData) For(idDossier ds.IdDossier) Events {
	raws := ld.events[idDossier]
	out := make([]Event, 0, len(raws))
	for _, event := range raws {
		out = append(out, ld.eventsContent.build(event, raws))
	}
	return out
}

// LoadEvent is a convenience method to load one [Event]
func LoadEvent(db evs.DB, id evs.IdEvent) (Event, error) {
	event, err := evs.SelectEvent(db, id)
	if err != nil {
		return Event{}, utils.SQLError(err)
	}
	content, err := loadEventsContent(db, id)
	if err != nil {
		return Event{}, utils.SQLError(err)
	}
	return content.build(event, nil), nil
}
