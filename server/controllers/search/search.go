package search

import (
	"cmp"
	"slices"
	"strings"

	cps "registro/sql/camps"
	pr "registro/sql/personnes"
	"registro/utils"
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
		s = utils.Normalize(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return Query{out}
}

// match returns true if [str] matches the query.
func (q Query) match(str string) bool {
	str = utils.Normalize(str)
	for _, r := range q.patterns {
		if !strings.Contains(str, r) {
			return false
		}
	}
	return true
}

type filterable interface{ pr.Personne | cps.Camp }

// QueryMatch returns true if [v] passes the query [qu]
func QueryMatch[T filterable](qu Query, v T) bool { return qu.match(stringify(v)) }

func stringify[T filterable](v T) string {
	switch v := any(v).(type) {
	case pr.Personne:
		return v.NomPrenom()
	case cps.Camp:
		return v.Nom + " " + v.DateDebut.String()
	default:
		panic("exhaustive switch")
	}
}

// FilterPersonnes ne se retreint pas automatiquement aux personnes non temporaires
func FilterPersonnes(list pr.Personnes, pattern string) (out []PersonneHeader) {
	rs := NewQuery(pattern)

	for _, v := range list {
		if QueryMatch(rs, v) {
			out = append(out, newPersonneHeader(v))
		}
	}

	slices.SortFunc(out, func(a, b PersonneHeader) int { return cmp.Compare(a.Label, b.Label) })

	return out
}

func FilterCamps(list cps.Camps, pattern string) (out []cps.Camp) {
	rs := NewQuery(pattern)

	for _, v := range list {
		if QueryMatch(rs, v) {
			out = append(out, v)
		}
	}

	return out
}
