package search

import (
	"slices"
	"strings"

	pr "registro/sql/personnes"
	"registro/utils"
)

const (
	poidsNom           = 3
	poidsPrenom        = 2
	poidsMail          = 4
	poidsDateNaissance = 3
	poidsSexe          = 1
)

// PatternsSimilarite exposes the fields used to search
// existing profils
type PatternsSimilarite struct {
	Nom           string
	Prenom        string
	Sexe          pr.Sexe
	DateNaissance pr.Date
	Mail          string
}

func (in *PatternsSimilarite) normalize() {
	in.Nom = strings.TrimSpace(strings.ToLower(in.Nom))
	in.Prenom = strings.TrimSpace(strings.ToLower(in.Prenom))
	in.Mail = strings.TrimSpace(strings.ToLower(in.Mail))
}

func (in PatternsSimilarite) scoreMax() (scoreMax int) {
	if in.Nom != "" {
		scoreMax += poidsNom
	}
	if in.Prenom != "" {
		scoreMax += poidsPrenom
	}
	if in.Mail != "" {
		scoreMax += poidsMail
	}
	if in.Sexe != 0 {
		scoreMax += poidsSexe
	}
	if !in.DateNaissance.Time().IsZero() {
		scoreMax += poidsDateNaissance
	}
	return scoreMax
}

// SelectAllPatternSimilaires ne charge que les champs requis
// par `PatternsSimilarite`
func SelectAllPatternSimilaires(db pr.DB) ([]pr.Personne, error) {
	// Champs utilisés par la recherche de profil : on évite de charger tous les champs
	const query = "SELECT Id, Nom, Prenom, Sexe, DateNaissance, Mail, IsTemp FROM personnes"

	rows, err := db.Query(query)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	defer rows.Close()

	var personnes []pr.Personne
	for rows.Next() {
		var item pr.Personne
		err := rows.Scan(
			&item.Id,
			&item.Nom,
			&item.Prenom,
			&item.Sexe,
			&item.DateNaissance,
			&item.Mail,
			&item.IsTemp,
		)
		if err != nil {
			return nil, utils.SQLError(err)
		}
		personnes = append(personnes, item)
	}
	if err := rows.Err(); err != nil {
		return nil, utils.SQLError(err)
	}

	return personnes, nil
}

// Enlève les accents de full et renvoie true si subs est dans full
// Renvoi false si une des deux chaines est vide.
func isIn(full, substr string) bool {
	return substr != "" && full != "" && strings.Contains(
		utils.Normalize(full), substr)
}

func comparaison(p pr.Personne, in PatternsSimilarite) (score int) {
	if in.Mail != "" && strings.Contains(strings.TrimSpace(p.Mail), in.Mail) {
		score += poidsMail
	}
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

type ScoredPersonne struct {
	Score    int
	Personne pr.Personne
}

// ChercheSimilaires renvoie les profils similaires à [in].
// Les personnes temporaires sont ignorées, puisque que l'on ne
// veut pas fusionner un profil entrant à un profil temporaire.
func ChercheSimilaires(personnes []pr.Personne, in PatternsSimilarite) (scoreMax int, out []ScoredPersonne) {
	const seuilRechercheSimilaire = 2

	scoreMax = in.scoreMax()
	if scoreMax == 0 {
		return
	}
	in.normalize()

	out = make([]ScoredPersonne, 0, 20)
	for _, p := range personnes {
		if p.IsTemp {
			continue
		}
		score := comparaison(p, in)
		if score >= seuilRechercheSimilaire {
			out = append(out, ScoredPersonne{score, p})
		}
	}

	slices.SortFunc(out, func(a, b ScoredPersonne) int { return b.Score - a.Score }) // decroissant

	return scoreMax, out
}
