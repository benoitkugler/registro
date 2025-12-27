package search

import (
	"slices"
	"strings"
	"unicode"

	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"golang.org/x/text/runes"
	"golang.org/x/text/unicode/rangetable"
)

var alphaNum = rangetable.Merge(unicode.L, unicode.Digit)

// Normalize remove accents, white space and any non alphanumeric caracters
func Normalize(s string) string {
	b := []byte(s)
	b = utils.RemoveAccents(b)
	b = runes.Remove(runes.NotIn(alphaNum)).Bytes(b)
	b = runes.Map(unicode.ToLower).Bytes(b)
	return string(b)
}

const (
	poidsNom           = 3
	poidsPrenom        = 2
	poidsDateNaissance = 3
	poidsSexe          = 1
)

// PatternsSimilarite exposes the fields used to search
// existing profils.
type PatternsSimilarite struct {
	Nom           string
	Prenom        string
	DateNaissance shared.Date

	Sexe pr.Sexe // if empty, the matching is done ignoring Sexe
}

// NewPatternsSimilarite selects the proper fields.
func NewPatternsSimilarite(pr pr.Etatcivil) PatternsSimilarite {
	return PatternsSimilarite{
		Nom:           pr.Nom,
		Prenom:        pr.Prenom,
		Sexe:          pr.Sexe,
		DateNaissance: pr.DateNaissance,
	}
}

func (ps *PatternsSimilarite) Personne() pr.Etatcivil {
	return pr.Etatcivil{
		Nom:           ps.Nom,
		Prenom:        ps.Prenom,
		Sexe:          ps.Sexe,
		DateNaissance: ps.DateNaissance,
	}
}

// match returns true if the pattern precisely matches
// the given [candidate], that is if all the fields matches,
// using normalization.
// [ps] should have been normalized beforehand.
func (ps *PatternsSimilarite) match(candidate pr.Etatcivil) bool {
	return ps.Nom == Normalize(candidate.Nom) &&
		ps.Prenom == Normalize(candidate.Prenom) &&
		(ps.Sexe == pr.NoSexe || ps.Sexe == candidate.Sexe) &&
		ps.DateNaissance == candidate.DateNaissance
}

func (in *PatternsSimilarite) normalize() {
	in.Nom = Normalize(in.Nom)
	in.Prenom = Normalize(in.Prenom)
}

func (in PatternsSimilarite) scoreMax() (scoreMax int) {
	if in.Nom != "" {
		scoreMax += poidsNom
	}
	if in.Prenom != "" {
		scoreMax += poidsPrenom
	}
	if in.Sexe != 0 {
		scoreMax += poidsSexe
	}
	if !in.DateNaissance.Time().IsZero() {
		scoreMax += poidsDateNaissance
	}
	return scoreMax
}

// Enlève les accents de full et renvoie true si subs est dans full
// Renvoi false si une des deux chaines est vide.
func isIn(full, substr string) bool {
	return substr != "" && full != "" && strings.Contains(
		Normalize(full), substr)
}

func comparaison(p pr.Personne, in PatternsSimilarite) (score int) {
	if !p.DateNaissance.Time().IsZero() && in.DateNaissance == p.DateNaissance {
		score += poidsDateNaissance
	}
	if p.Sexe != 0 && p.Sexe == in.Sexe {
		score += poidsSexe
	}
	if isIn(p.Nom, in.Nom) {
		score += poidsNom
	}
	if isIn(p.Prenom, in.Prenom) {
		score += poidsPrenom
	}
	return score
}

type PersonneHeader struct {
	Id            pr.IdPersonne
	Label         string
	Sexe          pr.Sexe
	DateNaissance shared.Date
	IsTemp        bool
}

func NewPersonneHeader(p pr.Personne) PersonneHeader {
	return PersonneHeader{
		p.Id,
		p.PrenomNOM(),
		p.Sexe,
		p.DateNaissance,
		p.IsTemp,
	}
}

// Match vérifie si [in] est déjà présent dans la liste [personnes],
// et renvoie le premier profil correspondant.
//
// Voir aussi [SelectAllFieldsForSimilaires]
func Match(personnes pr.Personnes, in PatternsSimilarite) (pr.IdPersonne, bool) {
	in.normalize()
	for _, personne := range personnes {
		if in.match(personne.Etatcivil) {
			return personne.Id, true
		}
	}
	return 0, false
}

type ScoredPersonne struct {
	ScorePercent int // between 0 and 100

	Personne PersonneHeader
}

// ChercheSimilaires renvoie les profils similaires à [in],
// triés par pertinence (meilleur en premier).
//
// Voir aussi [SelectAllFieldsForSimilaires]
func ChercheSimilaires(personnes pr.Personnes, in PatternsSimilarite) (scoreMax int, out []ScoredPersonne) {
	const seuilRechercheSimilaire = 2

	scoreMax = in.scoreMax()
	if scoreMax == 0 {
		return
	}
	in.normalize()

	out = make([]ScoredPersonne, 0, 8)
	for _, p := range personnes {
		if score := comparaison(p, in); score >= seuilRechercheSimilaire {
			out = append(out, ScoredPersonne{
				100 * score / scoreMax,
				NewPersonneHeader(p),
			})
		}
	}

	slices.SortFunc(out, func(a, b ScoredPersonne) int { return b.ScorePercent - a.ScorePercent }) // decroissant

	return scoreMax, out
}

// SelectAllFieldsForSimilaires ne charge que les champs requis
// par `PatternsSimilarite`.
// Les personnes temporaires sont ignorées, puisque que l'on ne
// veut pas fusionner un profil entrant à un profil temporaire.
//
// Errors are wrapped with [utils.SQLError]
func SelectAllFieldsForSimilaires(db pr.DB) (pr.Personnes, error) {
	// Champs utilisés par la recherche de profil : on évite de charger tous les champs
	const query = "SELECT Id, Nom, Prenom, Sexe, DateNaissance FROM personnes WHERE IsTemp IS False;"

	rows, err := db.Query(query)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	defer rows.Close()

	personnes := make(pr.Personnes)
	for rows.Next() {
		var item pr.Personne
		err := rows.Scan(
			&item.Id,
			&item.Nom,
			&item.Prenom,
			&item.Sexe,
			&item.DateNaissance,
		)
		if err != nil {
			return nil, utils.SQLError(err)
		}
		personnes[item.Id] = item
	}
	if err := rows.Err(); err != nil {
		return nil, utils.SQLError(err)
	}

	return personnes, nil
}
