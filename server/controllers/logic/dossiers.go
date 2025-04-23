package logic

import (
	"slices"
	"time"

	filesAPI "registro/controllers/files"
	"registro/crypto"
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

// LoadDossier is a convenient wrapper around [LoadDossiers]
func LoadDossier(db ds.DB, id ds.IdDossier) (Dossier, error) {
	ld, err := LoadDossiers(db, id)
	if err != nil {
		return Dossier{}, err
	}
	return ld.For(id), nil
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
	ld, err := LoadEventsByDossiers(db, ids...)
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

// ParticipantsExt is sorted by Camp then by Id
func (de *Dossier) ParticipantsExt() []cps.ParticipantCamp {
	ps := make([]cps.ParticipantCamp, 0, len(de.Participants))
	for _, part := range de.Participants {
		ps = append(ps, cps.ParticipantCamp{
			Camp: de.camps[part.IdCamp],
			ParticipantPersonne: cps.ParticipantPersonne{
				Participant: part,
				Personne:    de.personnesM[part.IdPersonne],
			},
		})
	}

	slices.SortFunc(ps, func(a, b cps.ParticipantCamp) int { return int(a.Participant.Id - b.Participant.Id) })
	slices.SortStableFunc(ps, func(a, b cps.ParticipantCamp) int { return int(a.Camp.Id - b.Camp.Id) })
	return ps
}

// Personnes returns the responsable first, then sort by ID
func (de *Dossier) Personnes() (out []pr.Personne) {
	out = append(out, de.Responsable())
	for _, part := range de.Participants {
		out = append(out, de.personnesM[part.IdPersonne])
	}
	slices.SortFunc(out[1:], func(a, b pr.Personne) int { return int(a.Id - b.Id) })
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
		if eventT := event.Created; eventT.After(last) {
			last = eventT
		}
	}
	return last
}

// DossierExt is the public version of a [Dossier],
// with all information resolved.
type DossierExt struct {
	Dossier      ds.Dossier
	Responsable  string
	Participants []cps.ParticipantCamp
	Aides        map[cps.IdParticipant]cps.Aides
	AidesFiles   map[cps.IdAide]filesAPI.PublicFile // optionnel

	Events    Events
	Paiements ds.Paiements

	Bilan BilanFinancesPub
}

type BilanParticipantPub struct {
	BilanParticipant

	AvecAides string
	Net       string
}

func (bp BilanParticipant) publish(taux ds.Taux) BilanParticipantPub {
	out := BilanParticipantPub{BilanParticipant: bp}

	out.AvecAides = taux.Convertible(bp.prixSansRemises(taux)).String()
	out.Net = taux.Convertible(bp.net(taux)).String()
	return out
}

type BilanFinancesPub struct {
	Inscrits map[cps.IdParticipant]BilanParticipantPub

	Demande string // total des participants, aides déjà déduises
	Aides   string // total des aides extérieures
	Recu    string // total des paiements
	Restant string // Demande - Recu
	Statut  StatutPaiement
}

func (d DossierFinance) Publish(key crypto.Encrypter) DossierExt {
	taux := d.taux
	b := d.Bilan()
	inscrits := make(map[cps.IdParticipant]BilanParticipantPub, len(b.inscrits))
	for k, v := range b.inscrits {
		inscrits[k] = v.publish(taux)
	}

	bilan := BilanFinancesPub{
		inscrits,
		taux.Convertible(ds.Montant{Cent: b.demande, Currency: b.currency}).String(),
		taux.Convertible(ds.Montant{Cent: b.aides, Currency: b.currency}).String(),
		taux.Convertible(ds.Montant{Cent: b.recu, Currency: b.currency}).String(),
		taux.Convertible(b.ApresPaiement()).String(),
		b.StatutPaiement(),
	}

	aideFiles := make(map[cps.IdAide]filesAPI.PublicFile)
	for _, l := range d.aides {
		for _, aide := range l {
			if file, ok := d.aidesFiles[aide.Id]; ok {
				aideFiles[aide.Id] = filesAPI.NewPublicFile(key, file)
			}
		}
	}
	return DossierExt{d.Dossier.Dossier, d.Responsable().PrenomNOM(), d.ParticipantsExt(), d.aides, aideFiles, d.Events, d.paiements, bilan}
}
