package personnes

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	reSepTel    = regexp.MustCompile("[ -/;\t]")
	reSepPrenom = regexp.MustCompile("[ -.]")
)

func upperFirst(s string) string {
	s = strings.ToLower(s)
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func formatPrenom(s string) string {
	parts := reSepPrenom.Split(string(s), -1)
	var tmp []string
	for _, p := range parts {
		if p != "" {
			tmp = append(tmp, upperFirst(p))
		}
	}
	return strings.Join(tmp, "-")
}

func (s Sexe) String() string {
	switch s {
	case Woman:
		return "Femme"
	case Man:
		return "Homme"
	default:
		return ""
	}
}

// Accord returns "e" for women
func (s Sexe) Accord() string {
	if s == Woman {
		return "e"
	}
	return ""
}

// StripTel return the number without spaces or delimiters
func StripTel(t string) string { return reSepTel.ReplaceAllString(t, "") }

func renderTels(t Tels, outerSep string) string {
	return strings.Join(t.NonEmpty(), outerSep)
}

func (t Tels) NonEmpty() []string {
	// reuse t buffer is OK because we just copied the array
	out := t[:0]
	if n := t[0]; n != "" {
		out = append(out, n)
	}
	if n := t[1]; n != "" {
		out = append(out, n)
	}
	return out
}

func (t Tels) String() string { return renderTels(t, ";") }

// StringLines renvoie une chaine sur plusieurs lignes, au format HTML
func (t Tels) StringHTML() string { return renderTels(t, "<br/>") }
