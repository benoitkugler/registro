package camps

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"

	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	sh "registro/sql/shared"
	"registro/utils"

	"golang.org/x/text/runes"
	"golang.org/x/text/unicode/rangetable"
)

type ParticipantCamp struct {
	Camp Camp
	ParticipantPersonne
}

type ParticipantPersonne struct {
	Participant Participant
	Personne    pr.Personne
}

// StatistiquesInscrits détails le nombre d'inscriptions
// sur un séjour
type StatistiquesInscrits struct {
	Inscriptions                            int // nombre total
	InscriptionsFilles, InscriptionsSuisses int
	Valides, ValidesFilles, ValidesSuisses  int // confirmés
	Refus, AStatuer, Exceptions, Attente    int
}

func (stats *StatistiquesInscrits) add(p ParticipantPersonne) {
	stats.Inscriptions += 1

	isFille := p.Personne.Sexe == pr.Woman
	isSuisse := p.Personne.Nationnalite.IsSuisse

	if isFille {
		stats.InscriptionsFilles += 1
	}
	if isSuisse {
		stats.InscriptionsSuisses += 1
	}

	switch p.Participant.Statut {
	case Refuse:
		stats.Refus += 1
	case AStatuer:
		stats.AStatuer += 1
	case AttenteProfilInvalide:
		stats.Exceptions += 1
	case AttenteCampComplet, EnAttenteReponse:
		stats.Attente += 1
	case Inscrit:
		stats.Valides += 1
		if isFille {
			stats.ValidesFilles += 1
		}
		if isSuisse {
			stats.ValidesSuisses += 1
		}
	}
}

type CampsData struct {
	Camps        Camps
	participants map[IdCamp]Participants // covering [camps]
	personnes    pr.Personnes            // covering [camps]
}

// LoadCamps loads the given camps and their participants,
// and wraps the error.
func LoadCamps(db DB, ids []IdCamp) (CampsData, error) {
	camps, err := SelectCamps(db, ids...)
	if err != nil {
		return CampsData{}, utils.SQLError(err)
	}
	participants, err := SelectParticipantsByIdCamps(db, ids...)
	if err != nil {
		return CampsData{}, utils.SQLError(err)
	}
	byCamp := participants.ByIdCamp()
	personnes, err := pr.SelectPersonnes(db, participants.IdPersonnes()...)
	if err != nil {
		return CampsData{}, utils.SQLError(err)
	}
	return CampsData{camps, byCamp, personnes}, nil
}

func (cps CampsData) For(id IdCamp) CampData {
	return CampData{cps.Camps[id], cps.participants[id], cps.personnes}
}

func (cps CampsData) IdDossiers() []ds.IdDossier {
	s := utils.NewSet[ds.IdDossier]()
	for _, ps := range cps.participants {
		s.AddAll(ps.IdDossiers())
	}
	return s.Keys()
}

// Personnes returns the [Personne]s for all [Participant]s
func (cps CampsData) Personnes(onlyInscrits bool) pr.Personnes {
	out := make(pr.Personnes)
	for _, ps := range cps.participants {
		for _, p := range ps {
			if onlyInscrits && p.Statut != Inscrit {
				continue
			}
			out[p.IdPersonne] = cps.personnes[p.IdPersonne]
		}
	}
	return out
}

// CampData permet d'accéder à diverses
// propriété d'un séjour nécessitant la liste des inscrits.
type CampData struct {
	Camp         Camp
	participants Participants // liste (exacte) des participants du séjour
	// Doit contenir au moins les participants
	personnes pr.Personnes
}

// LoadCamp is a convenient wrapper around [LoadCamps] for
// a single camp.
func LoadCamp(db DB, id IdCamp) (CampData, error) {
	out, err := LoadCamps(db, []IdCamp{id})
	if err != nil {
		return CampData{}, err
	}
	return out.For(id), nil
}

func (cd CampData) Participants(onlyInscrits bool) []ParticipantPersonne {
	out := make([]ParticipantPersonne, 0, len(cd.participants))
	for _, participant := range cd.participants {
		if onlyInscrits && participant.Statut != Inscrit {
			continue
		}
		pe := cd.personnes[participant.IdPersonne]
		out = append(out, ParticipantPersonne{participant, pe})
	}
	return out
}

func (cd CampData) IdDossiers() []ds.IdDossier {
	return cd.participants.IdDossiers()
}

// Personnes returns the [Personne]s for all [Participant]s
func (cd CampData) Personnes(onlyInscrits bool) pr.Personnes {
	out := make(pr.Personnes)
	for _, p := range cd.participants {
		if onlyInscrits && p.Statut != Inscrit {
			continue
		}
		out[p.IdPersonne] = cd.personnes[p.IdPersonne]
	}
	return out
}

func (cd CampData) Stats() StatistiquesInscrits {
	var stats StatistiquesInscrits
	for _, p := range cd.Participants(false) {
		stats.add(p)
	}
	return stats
}

// restePlace vaut `true` si l'ajout de participants
// ne dépasse pas le nombre de places autorisées
func (cd *Camp) restePlace(stats StatistiquesInscrits, participants []pr.Personne) bool {
	current := stats.Valides
	return current+len(participants) <= cd.Places
}

// keepEquilibreGF renvoie `true` si l'ajout des [participants]
// ne perturbe pas l'équilibre G/F (ou si le séjour ne demande pas d'équilibre).
func (cd *Camp) keepEquilibreGF(stats StatistiquesInscrits, participants []pr.Personne) bool {
	if !cd.NeedEquilibreGF {
		return true
	}
	var newG, newF int
	for _, p := range participants {
		if p.Sexe == pr.Woman {
			newF += 1
		} else {
			newG += 1
		}
	}
	currentG, currentF := stats.Valides-stats.ValidesFilles, stats.ValidesFilles
	// on utilise l'heuristique suivante :
	// dépasser 60% des places prévues détruit l'équilibre
	seuil := cd.Places * 6 / 10
	return currentG+newG <= seuil && currentF+newF <= seuil
}

// StatutCauses expose une série de critère
// de validité pour l'inscription d'un participant à un camp,
// ainsi que le statut conseillé
type StatutCauses struct {
	AgeMin, AgeMax, EquilibreGF, Place bool
}

// Hint indique comment placer le participant
func (s StatutCauses) Hint() StatutParticipant {
	if !(s.AgeMin && s.AgeMax) {
		return AttenteProfilInvalide
	} else if !(s.Place && s.EquilibreGF) {
		return AttenteCampComplet
	}
	return Inscrit
}

// Status détermine la validité de l'inscription des personnes
// données par [participants], renvoyant une liste de la même longueur
func (cd CampData) Status(participants []pr.Personne) []StatutCauses {
	stats := cd.Stats()

	restePlace := cd.Camp.restePlace(stats, participants)
	equilibreGF := cd.Camp.keepEquilibreGF(stats, participants)

	out := make([]StatutCauses, len(participants))
	for i, part := range participants {
		isMinValid, isMaxValid := cd.Camp.IsAgeValide(part.DateNaissance)
		out[i] = StatutCauses{
			AgeMin:      isMinValid,
			AgeMax:      isMaxValid,
			Place:       restePlace,
			EquilibreGF: equilibreGF,
		}
	}
	return out
}

// Label renvoie une description courte : Nom Année
func (c Camp) Label() string {
	return fmt.Sprintf("%s %d", c.Nom, c.DateDebut.Time().Year())
}

func (cp *Camp) Plage() sh.Plage { return sh.Plage{From: cp.DateDebut, Duree: cp.Duree} }

func (cp *Camp) DateFin() sh.Date { return cp.Plage().To() }

// IsPassedBy renvoie `true` si le séjour est
// passé d'au moins [jours].
func (cp *Camp) IsPassedBy(jours int) bool {
	const oneDay = 24 * time.Hour
	dateFin := cp.DateFin().Time()
	return time.Now().After(dateFin.Add(time.Duration(jours) * oneDay))
}

// AgeDebutCamp renvoie l'âge qu'aura une personne née le 'dateNaissance' au premier jour
// du séjour.
func (cp *Camp) AgeDebutCamp(dateNaissance sh.Date) int {
	return dateNaissance.Age(cp.DateDebut.Time())
}

// IsAgeValide renvoie le statut correspondant aux âges min et max du séjour
func (cp *Camp) IsAgeValide(dateNaissance sh.Date) (min, max bool) {
	age := cp.AgeDebutCamp(dateNaissance)
	min = age >= cp.AgeMin
	max = age <= cp.AgeMax
	return min, max
}

var alphaNumSpace = rangetable.Merge(unicode.L, unicode.Digit, unicode.Zs)

// Slug returns a URL compatible string used as ID
func (cp *Camp) Slug() string {
	b := utils.RemoveAccents([]byte(cp.Nom))
	b = runes.Remove(runes.NotIn(alphaNumSpace)).Bytes(b)
	name := strings.ReplaceAll(string(b), " ", "-")
	name = strings.ToLower(name)
	slug := fmt.Sprintf("%s-%d", name, cp.DateDebut.Time().Year())
	return url.QueryEscape(slug)
}

// Check assure la validité de divers champs.
func (c *Camp) Check() error {
	if c.Duree < 1 {
		return errors.New("invalid Duree")
	}
	if c.DateDebut.Time().Year() < 2020 {
		return errors.New("invalid DateDebut")
	}
	if c.Places < 1 {
		return errors.New("invalid Places")
	}
	if c.AgeMin < 0 {
		return errors.New("invalid AgeMin")
	}
	if c.AgeMax < 1 || c.AgeMax < c.AgeMin {
		return errors.New("invalid AgeMax")
	}
	if c.Prix.Cent < 0 {
		return errors.New("invalid Prix")
	}
	for _, perc := range c.OptionQuotientFamilial {
		if !(0 <= perc && perc <= 100) {
			return errors.New("invalid OptionQuotientFamilial percentage")
		}
	}
	if c.OptionPrix.Active == PrixJour {
		if c.Duree != len(c.OptionPrix.Jours) {
			return errors.New("invalid OptionPrix.Jour length")
		}
	}
	if c.OptionPrix.Active == PrixStatut {
		if len(c.OptionPrix.Statuts) == 0 {
			return errors.New("invalid OptionPrix.Status length")
		}
	}
	return nil
}

type CampExt struct {
	Camp Camp
	// IsTerminated is 'true' when the camp
	// is over by (at least) 1 day, even if the 'Ouvert' tag is still on.
	IsTerminated bool
	Slug         string
}

func (cp Camp) Ext() CampExt {
	return CampExt{cp, cp.IsPassedBy(1), cp.Slug()}
}

// TrouveGroupe cherche parmis les groupes possibles celui qui pourrait convenir.
// Normalement, les groupes respectent un invariant de continuité sur les plages,
// imposé par le frontend.
// Si plusieurs pourraient convenir, un seul est renvoyé, de façon arbitraire.
func (gs Groupes) TrouveGroupe(dateNaissance shared.Date) (Groupe, bool) {
	for _, g := range gs {
		if g.Plage.Contains(dateNaissance) {
			return g, true
		}
	}
	// on a trouvé aucun groupe
	return Groupe{}, false
}

// Directeur renvoie le directeur (unique par construction)
// du séjour donné, où false s'il n'existe pas.
func (equipiers Equipiers) Directeur() (Equipier, bool) {
	for _, eq := range equipiers {
		if eq.Roles.Is(Direction) {
			return eq, true
		}
	}
	return Equipier{}, false
}

// Direction renvoie les équipiers dans la direction ou sous-direction.
// S'il existe, le directeur est en premier
func (equipiers Equipiers) Direction() []Equipier {
	var direction, adjoints []Equipier
	for _, eq := range equipiers {
		if eq.Roles.Is(Direction) {
			direction = append(direction, eq)
		} else if eq.Roles.Is(Adjoint) {
			adjoints = append(adjoints, eq)
		}
	}
	slices.SortFunc(direction, func(a, b Equipier) int { return int(a.Id - b.Id) })
	slices.SortFunc(adjoints, func(a, b Equipier) int { return int(a.Id - b.Id) })

	return append(direction, adjoints...)
}

// Resolve renvoie :
//   - le montant dans le cas d'une aide absolue
//   - le montant fois le nombre de jours (en prenant en compte une éventuelle limite) sinon
func (ai Aide) Resolve(nbJours int) Montant {
	val := ai.Valeur
	if ai.ParJour {
		limite := ai.NbJoursMax
		if limite > 0 && limite < nbJours { // apply the limit
			nbJours = limite
		}
		val.Cent = val.Cent * nbJours
	}
	return val
}
