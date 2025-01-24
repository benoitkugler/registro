package search

import (
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

type Filterable interface{ pr.Personne | cps.Camp }

func stringify[T Filterable](v T) string {
	switch v := any(v).(type) {
	case pr.Personne:
		return v.NomPrenom()
	case cps.Camp:
		return v.Nom + " " + v.DateDebut.String()
	default:
		panic("exhaustive switch")
	}
}

func Filter[T Filterable](list []T, pattern string) (out []T) {
	rs := normalizeSearch(pattern)

	for _, v := range list {
		if matchAll(rs, stringify(v)) {
			out = append(out, v)
		}
	}

	return out
}
