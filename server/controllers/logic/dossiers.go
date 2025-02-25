package logic

import (
	"slices"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"
)

// Dossiers stores enough data to
// handle most of the [Dossier] related logic.
type Dossiers struct {
	Dossiers        ds.Dossiers
	allParticipants []cps.IdParticipant
	participants    map[ds.IdDossier]cps.Participants
	personnes       pr.Personnes
	camps           cps.Camps
	events          EventsData
}

// LoadDossiers wraps the SQL error
func LoadDossiers(db ds.DB, ids ...ds.IdDossier) (Dossiers, error) {
	dossiers, err := ds.SelectDossiers(db, ids...)
	if err != nil {
		return Dossiers{}, utils.SQLError(err)
	}

	// select the participants and associated people
	links, err := cps.SelectParticipantsByIdDossiers(db, ids...)
	if err != nil {
		return Dossiers{}, utils.SQLError(err)
	}
	participants := links.ByIdDossier()

	personnes, err := pr.SelectPersonnes(db, append(dossiers.IdResponsables(), links.IdPersonnes()...)...)
	if err != nil {
		return Dossiers{}, utils.SQLError(err)
	}

	// load the camps
	camps, err := cps.SelectCamps(db, links.IdCamps()...)
	if err != nil {
		return Dossiers{}, utils.SQLError(err)
	}
	// load the messages
	ld, err := LoadEvents(db, ids...)
	if err != nil {
		return Dossiers{}, utils.SQLError(err)
	}

	return Dossiers{dossiers, links.IDs(), participants, personnes, camps, ld}, nil
}

func (ld *Dossiers) For(id ds.IdDossier) Dossier {
	dossier, participants := ld.Dossiers[id], ld.participants[id]
	events := ld.events.For(id)
	return Dossier{dossier, participants, ld.personnes, ld.camps, events}
}

type Dossier struct {
	Dossier      ds.Dossier
	Participants cps.Participants // Liste exacte
	personnesM   pr.Personnes     // containing at least the reponsable and participants
	camps        cps.Camps        // containing at least the camps for [participants]
	Events       Events
}

func (de *Dossier) Responsable() pr.Personne { return de.personnesM[de.Dossier.IdResponsable] }

// ParticipantsExt is sorted by Id
func (de *Dossier) ParticipantsExt() []cps.ParticipantExt {
	ps := make([]cps.ParticipantExt, 0, len(de.Participants))
	for _, part := range de.Participants {
		ps = append(ps, cps.ParticipantExt{
			Participant: part,
			Camp:        de.camps[part.IdCamp],
			Personne:    de.personnesM[part.IdPersonne],
		})
	}

	slices.SortFunc(ps, func(a, b cps.ParticipantExt) int { return int(a.Participant.Id - b.Participant.Id) })
	return ps
}

// Personnes returns the responsable first
func (de *Dossier) Personnes() (out []pr.Personne) {
	out = append(out, de.Responsable())
	for _, part := range de.Participants {
		out = append(out, de.personnesM[part.IdPersonne])
	}
	return out
}

// PersonnesFor returns the personne for every given participants
func (de *Dossier) PersonnesFor(participants []cps.Participant) []pr.Personne {
	out := make([]pr.Personne, len(participants))
	for index, part := range participants {
		out[index] = de.personnesM[part.IdPersonne]
	}
	return out
}

// Camps returns the map of [Camp]s concerned by this dossier.
func (de *Dossier) Camps() cps.Camps {
	out := make(cps.Camps)
	for _, part := range de.Participants {
		out[part.IdCamp] = de.camps[part.IdCamp]
	}
	return out
}

// Time returns the last interaction in the message track
func (de *Dossier) Time() time.Time {
	last := de.Dossier.MomentInscription // start with the inscription
	for _, event := range de.Events {
		if eventT := event.Event.Created; eventT.After(last) {
			last = eventT
		}
	}
	return last
}
