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

// matchAll returns true if [str] contains all [patterns]
func matchAll(patterns []string, str string) bool {
	str = utils.Normalize(str)
	for _, r := range patterns {
		if !strings.Contains(str, r) {
			return false
		}
	}
	return true
}

func normalizeSearch(pattern string) (out []string) {
	// special case for no filter
	if pattern == "*" {
		return out
	}

	for _, s := range strings.Fields(pattern) {
		s = utils.Normalize(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

type filterable interface{ pr.Personne | cps.Camp }

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
	rs := normalizeSearch(pattern)

	for _, v := range list {
		if matchAll(rs, stringify(v)) {
			out = append(out, newPersonneHeader(v))
		}
	}

	slices.SortFunc(out, func(a, b PersonneHeader) int { return cmp.Compare(a.Label, b.Label) })

	return out
}

func FilterCamps(list cps.Camps, pattern string) (out []cps.Camp) {
	rs := normalizeSearch(pattern)

	for _, v := range list {
		if matchAll(rs, stringify(v)) {
			out = append(out, v)
		}
	}

	return out
}
