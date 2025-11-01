package search

import (
	"cmp"
	"slices"
	"strings"

	pr "registro/sql/personnes"
)

// Fonctions de recherche rapide (par string)

// Query stores a "fuzzy" search
type Query struct {
	patterns []string
}

// NewQuery normalizes the given pattern
// Note that an empty [pattern] matches everything
func NewQuery(pattern string) Query {
	var out []string
	for _, s := range strings.Fields(pattern) {
		s = Normalize(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return Query{out}
}

// QueryMatch returns true if [v] passes the query [qu],
// that is if every chunk matches.
func (qu Query) Match(v pr.Personne) bool {
	str := v.NOMPrenom()
	str = Normalize(str)
	for _, r := range qu.patterns {
		if !strings.Contains(str, r) {
			return false
		}
	}
	return true
}

// FilterPersonnes ne se retreint pas automatiquement aux personnes non temporaires
func FilterPersonnes(list pr.Personnes, pattern string) (out []PersonneHeader) {
	rs := NewQuery(pattern)

	for _, v := range list {
		if rs.Match(v) {
			out = append(out, NewPersonneHeader(v))
		}
	}

	slices.SortFunc(out, func(a, b PersonneHeader) int { return int(a.Id - b.Id) })
	slices.SortStableFunc(out, func(a, b PersonneHeader) int { return cmp.Compare(a.Label, b.Label) })

	return out
}
