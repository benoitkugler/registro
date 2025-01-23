package personnes

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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

func (d Date) String() string {
	da := time.Time(d)
	if da.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d", da.Day(), da.Month(), da.Year())
}

const innerTelSep = " "

func (t Tel) String() string { return formatTelSep(string(t), innerTelSep) }

// stripTel return the number without spaces or delimiters
func stripTel(t string) string { return reSepTel.ReplaceAllString(t, "") }

func formatTelSep(t string, separator string) string {
	t = stripTel(t)
	if len(t) < 8 {
		return t
	}
	start := len(t) - 8
	chunks := []string{t[:start]}
	for i := 0; i < 4; i++ {
		chunks = append(chunks, t[start+2*i:start+2*i+2])
	}
	return strings.Join(chunks, separator)
}

func renderTels(t Tels, innerSep, outerSep string) string {
	fmted := make([]string, len(t))
	for index, tel := range t {
		fmted[index] = formatTelSep(tel, innerSep)
	}
	return strings.Join(fmted, outerSep)
}

func (t Tels) String() string { return renderTels(t, innerTelSep, ";") }
