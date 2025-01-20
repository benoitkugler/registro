package utils

import "math/rand"

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
