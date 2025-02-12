package events

import (
	"slices"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"
)

// Event exposes on event on the dossier track
type Event struct {
	Event   events.Event
	Content Content
}

// Event exposes on event on the dossier track
type Content interface {
	isContent()
}

func (Supprime) isContent()        {}
func (Inscription) isContent()     {}
func (AccuseReception) isContent() {}
func (Message) isContent()         {}
func (Facture) isContent()         {}
func (CampDocs) isContent()        {}
func (PlaceLiberee) isContent()    {}
func (Attestation) isContent()     {}
func (Sondage) isContent()         {}

type (
	Supprime        struct{}
	Inscription     struct{}
	AccuseReception struct{}
)

type Message struct {
	Message          events.EventMessage
	OrigineCampLabel string   // optionnel
	VuParCamps       []string // labels
}

// m must have kind [Message]
func (ld *Loader) newMessage(ev events.Event) Message {
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
	CampLabel string
}

// m must have kind [CampDocs]
func (ld *Loader) newCampDocs(ev events.Event) CampDocs {
	m := ld.campDocs[ev.Id]
	camp := ld.camps[m.IdCamp]
	return CampDocs{CampLabel: camp.Label()}
}

type PlaceLiberee struct {
	ParticipantLabel string
	CampLabel        string
}

// m must have kind [PlaceLiberee]
func (ld *Loader) newPlaceLiberee(ev events.Event) PlaceLiberee {
	m := ld.placeLiberees[ev.Id]
	participant := ld.participants[m.IdParticipant]
	camp := ld.camps[participant.IdCamp]
	pers := ld.personnes[participant.IdPersonne]
	return PlaceLiberee{ParticipantLabel: pers.PrenomN(), CampLabel: camp.Label()}
}

type Attestation struct {
	Distribution events.Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool
}

// m must have kind [Attestation]
func (ld *Loader) newAttestation(ev events.Event) Attestation {
	m := ld.attestations[ev.Id]
	return Attestation{Distribution: m.Distribution, IsPresence: m.IsPresence}
}

type Sondage struct {
	CampLabel string
}

// m must have kind [Sondage]
func (ld *Loader) newSondage(ev events.Event) Sondage {
	m := ld.campDocs[ev.Id]
	camp := ld.camps[m.IdCamp]
	return Sondage{CampLabel: camp.Label()}
}

type Loader struct {
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

// NewLoaderFor loads the data required to build the events
// linked to the given dossiers.
//
// It does wrap any error encountered.
func NewLoaderFor(db events.DB, dossiers ...ds.IdDossier) (out Loader, _ error) {
	allEvents, err := events.SelectEventsByIdDossiers(db, dossiers...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.events = allEvents.ByIdDossier()
	ids := allEvents.IDs()

	tmp1, err := events.SelectEventMessagesByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.messages = tmp1.ByIdEvent()

	tmp1bis, err := events.SelectEventMessageVusByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.vupars = tmp1bis.ByIdEvent()

	tmp2, err := events.SelectEventCampDocssByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.campDocs = tmp2.ByIdEvent()

	tmp3, err := events.SelectEventPlaceLibereesByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.placeLiberees = tmp3.ByIdEvent()

	tmp4, err := events.SelectEventAttestationsByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.attestations = tmp4.ByIdEvent()

	tmp5, err := events.SelectEventSondagesByIdEvents(db, ids...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}
	out.sondages = tmp5.ByIdEvent()

	out.participants, err = cps.SelectParticipants(db, tmp3.IdParticipants()...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
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
		return Loader{}, utils.SQLError(err)
	}
	out.personnes, err = pr.SelectPersonnes(db, out.participants.IdPersonnes()...)
	if err != nil {
		return Loader{}, utils.SQLError(err)
	}

	return out, nil
}

func (ld *Loader) Events(idDossier ds.IdDossier) []Event {
	raws := ld.events[idDossier]
	out := make([]Event, len(raws))
	for i, event := range raws {
		out[i].Event = event
		switch event.Kind {
		case events.Supprime:
			out[i].Content = Supprime{}
		case events.Inscription:
			out[i].Content = Inscription{}
		case events.AccuseReception:
			out[i].Content = AccuseReception{}
		case events.Message:
			out[i].Content = ld.newMessage(event)
		case events.PlaceLiberee:
			out[i].Content = ld.newPlaceLiberee(event)
		case events.Facture:
			out[i].Content = Facture{}
		case events.CampDocs:
			out[i].Content = ld.newCampDocs(event)
		case events.Attestation:
			out[i].Content = ld.newAttestation(event)
		case events.Sondage:
			out[i].Content = ld.newSondage(event)
		}
	}
	return out
}
