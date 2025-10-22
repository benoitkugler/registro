package logic

import (
	"errors"
	"fmt"
	"html/template"
	"slices"
	"strings"
	"time"

	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"
)

// Dossiers stores enough data to
// handle most of the [Dossier] related logic.
type Dossiers struct {
	Dossiers              ds.Dossiers
	participantsIDs       []cps.IdParticipant
	participantsByDossier map[ds.IdDossier]cps.Participants
	personnes             pr.Personnes
	camps                 cps.Camps // all camps used by [participantsIDs]
	events                EventsData
}

// LoadDossier is a convenient wrapper around [LoadDossiers]
func LoadDossier(db ds.DB, id ds.IdDossier) (Dossier, error) {
	ld, err := LoadDossiers(db, id)
	if err != nil {
		return Dossier{}, err
	}
	if _, has := ld.Dossiers[id]; !has {
		return Dossier{}, errors.New("Dossier introuvable")
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
	dossier, participants := ld.Dossiers[id], ld.participantsByDossier[id]
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

// ParticipantsExtReal is like [ParticipantsExt], but is restricted to inscrits
// with started camp (at current time)
// It returns [true] if all the camps for inscrits have started and there is at least
// one inscrit.
func (de *Dossier) ParticipantsExtReal() ([]cps.ParticipantCamp, bool) {
	var (
		out        []cps.ParticipantCamp
		allStarted = true
	)
	for _, p := range de.ParticipantsExt() {
		if p.Participant.Statut != cps.Inscrit {
			continue
		}
		if hasStarted := p.Camp.DateDebut.Time().Before(time.Now()); !hasStarted {
			allStarted = false
			continue
		}
		out = append(out, p)
	}
	return out, allStarted && len(out) != 0
}

// Personnes returns the responsable first, then sort by ID.
// Among participants, repetition are removed
func (de *Dossier) Personnes() (out []pr.Personne) {
	out = append(out, de.Responsable())
	uniquesParticipants := utils.NewSet(de.Participants.IdPersonnes()...)
	for _, id := range utils.MapKeysSorted(uniquesParticipants) {
		out = append(out, de.personnesM[id])
	}
	return out
}

func (de *Dossier) ParticipantsLabels() string {
	personnes := de.Personnes()
	// extract participants
	chunks := make([]string, 0, len(personnes)-1)
	for _, pe := range personnes[1:] {
		chunks = append(chunks, pe.PrenomNOM())
	}
	return strings.Join(chunks, ", ")
}

// PersonnesFor returns the personne for every given participants
func (de *Dossier) PersonnesFor(participants []cps.Participant) []pr.Personne {
	out := make([]pr.Personne, len(participants))
	for index, part := range participants {
		out[index] = de.personnesM[part.IdPersonne]
	}
	return out
}

// PersonneFor returns the personne for the given participant
func (de *Dossier) PersonneFor(participant cps.Participant) pr.Personne {
	return de.personnesM[participant.IdPersonne]
}

// Camps returns the map of [Camp]s concerned by this dossier.
func (de *Dossier) Camps() cps.Camps { return de.campsInscrits(false) }

// Camps returns the map of [Camp]s concerned by this dossier,
// with at least one 'inscrit'.
func (de *Dossier) CampsInscrits() cps.Camps { return de.campsInscrits(true) }

func (de *Dossier) campsInscrits(onlyInscrits bool) cps.Camps {
	out := make(cps.Camps)
	for _, part := range de.Participants {
		if onlyInscrits && part.Statut != cps.Inscrit {
			continue
		}
		out[part.IdCamp] = de.camps[part.IdCamp]
	}
	return out
}

// FirstCampFor returns the first camp (defined by [DateDebut]) that [personne]
// will attend.
// It return false if [personne] is in waiting list for every camp.
func (de *Dossier) FirstCampFor(personne pr.IdPersonne) (cps.Camp, bool) {
	var camps []cps.Camp
	for _, participant := range de.Participants {
		if participant.IdPersonne == personne && participant.Statut == cps.Inscrit {
			camps = append(camps, de.camps[participant.IdCamp])
		}
	}
	if len(camps) == 0 {
		return cps.Camp{}, false
	}
	slices.SortFunc(camps, func(a, b cps.Camp) int { return a.DateDebut.Time().Compare(b.DateDebut.Time()) })
	return camps[0], true
}

// LastEventTime returns the last interaction in the events track
func (de *Dossier) LastEventTime() time.Time {
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
	AidesFiles   map[cps.IdAide]PublicFile // optionnel

	Events    Events
	Paiements ds.Paiements

	Bilan BilanFinancesPub
}

type BilanParticipantPub struct {
	BilanParticipant

	AvecAides string // mais sans remises
	Net       string // avec aides et remises
}

func (bp BilanParticipant) publish(taux ds.Taux) BilanParticipantPub {
	out := BilanParticipantPub{BilanParticipant: bp}

	out.AvecAides = taux.Convertible(bp.prixSansRemises(taux)).String()
	out.Net = taux.Convertible(bp.net(taux)).String()
	return out
}

// RemisesHTML returns a description of the remises,
// or an empty string
func (bp BilanParticipantPub) RemisesHTML() template.HTML {
	rem := bp.Remises
	if rem.ReducEnfants == 0 && rem.ReducEquipiers == 0 && rem.ReducSpeciale.Cent == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf("<i>Remise nombre d'enfants : %d%%    Remise équipiers : %d%%    Remise spéciale : %s</i>",
		rem.ReducEnfants, rem.ReducEquipiers, rem.ReducSpeciale))
}

type BilanFinancesPub struct {
	Inscrits map[cps.IdParticipant]BilanParticipantPub

	Demande string // total des participants, aides déjà déduises
	// Total des aides extérieures, ou vide
	Aides   string
	Recu    string // total des paiements
	Restant string // Demande - Recu
	Statut  StatutPaiement

	// Prix indicatif (sans remises ni aides) des prix
	// des séjours des participants non inscrits,
	// ou vide.
	DemandeEnAttenteValidation string
}

func (d DossierFinance) Publish(key crypto.Encrypter) DossierExt {
	taux := d.taux
	b := d.Bilan()
	inscrits := make(map[cps.IdParticipant]BilanParticipantPub, len(b.inscrits))
	for k, v := range b.inscrits {
		inscrits[k] = v.publish(taux)
	}

	enAttente := ""
	if b.demandeEnAttente != 0 {
		enAttente = taux.Convertible(ds.Montant{Cent: b.demandeEnAttente, Currency: b.currency}).String()
	}
	aides := ""
	if b.aides != 0 {
		aides = taux.Convertible(ds.Montant{Cent: b.aides, Currency: b.currency}).String()
	}

	bilan := BilanFinancesPub{
		inscrits,
		taux.Convertible(ds.Montant{Cent: b.demande, Currency: b.currency}).String(),
		aides,
		taux.Convertible(ds.Montant{Cent: b.recu, Currency: b.currency}).String(),
		taux.Convertible(b.ApresPaiement()).String(),
		b.StatutPaiement(),
		enAttente,
	}

	aideFiles := make(map[cps.IdAide]PublicFile)
	for _, l := range d.aides {
		for _, aide := range l {
			if file, ok := d.aidesFiles[aide.Id]; ok {
				aideFiles[aide.Id] = NewPublicFile(key, file)
			}
		}
	}
	return DossierExt{d.Dossier.Dossier, d.Responsable().PrenomNOM(), d.ParticipantsExt(), d.aides, aideFiles, d.Events, d.paiements, bilan}
}
