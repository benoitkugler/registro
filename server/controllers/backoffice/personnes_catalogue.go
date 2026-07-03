package backoffice

import (
	"strings"
)

// splits [adresse], ensuring len(adresse1) <= maxWidth
func formatAdresse(adresse_ string, maxWidth int) (adresse1, adresse2 string) {
	adresse := []rune(adresse_) // make sure we properly handle accents
	if len(adresse) <= maxWidth {
		return adresse_, ""
	}

	// 1: split at the first , (comma)
	adresse1, adresse2, _ = strings.Cut(adresse_, ",")
	adresse1, adresse2 = strings.TrimSpace(adresse1), strings.TrimSpace(adresse2)
	if len([]rune(adresse1)) <= maxWidth {
		return adresse1, adresse2
	}

	// 2: try to split at the last possible -
	if i := lastIndex(adresse[:maxWidth], '-'); i != -1 {
		return trimSpace(adresse[:i]), trimSpace(adresse[i+1:]) // do not include '-'
	}

	// 3: try to split by space
	if i := lastIndex(adresse[:maxWidth], ' '); i != -1 {
		return trimSpace(adresse[:i]), trimSpace(adresse[i:])
	}

	// default to split between words
	return string(adresse[:maxWidth]), string(adresse[maxWidth:])
}

func trimSpace(s []rune) string {
	return strings.TrimSpace(string(s))
}

// see https://github.com/golang/go/issues/63128
func lastIndex(s []rune, c rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
