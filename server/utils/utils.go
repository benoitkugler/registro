package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"unicode"

	"github.com/lib/pq"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	letterRunes  = []rune("azertyuiopqsdfghjklmwxcvbn123456789")
	specialRunes = []rune(" é @ ! ?&èïab ")
)

func RandString(n int, specialChars bool) string {
	b := make([]rune, n)
	props, maxLength := letterRunes, len(letterRunes)
	if specialChars {
		props = append(props, specialRunes...)
		maxLength += len(specialRunes)
	}
	for i := range b {
		b[i] = props[rand.Intn(maxLength)]
	}
	return string(b)
}

var noAccent = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func removeAccents(s []byte) []byte {
	output, _, err := transform.Bytes(noAccent, s)
	if err != nil {
		return s
	}
	return output
}

func Normalize(s string) string {
	return string(removeAccents(bytes.ToLower(bytes.TrimSpace([]byte(s)))))
}

func SQLError(err error) error {
	if err, ok := err.(*pq.Error); ok {
		return fmt.Errorf("La requête SQL (table %s) a échoué : %s", err.Table, err)
	}
	return fmt.Errorf("La requête SQL a échoué : %s %T", err, err)
}
