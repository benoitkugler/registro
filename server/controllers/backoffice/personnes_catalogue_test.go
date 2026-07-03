package backoffice

import (
	"testing"

	tu "registro/utils/testutils"
)

func Test_formatAdresse(t *testing.T) {
	tests := []struct {
		adresse  string
		maxWidth int
		want     string
		want2    string
	}{
		{"", 10, "", ""},
		{"", 0, "", ""},
		{"Résidence 3, 4 Allée", 30, "Résidence 3, 4 Allée", ""},
		{"Résidence 3 , 4 Allée", 11, "Résidence 3", "4 Allée"},
		{"Résidence 3, 4 Allée du bon lait", 20, "Résidence 3", "4 Allée du bon lait"},
		{"Bâtiment Normandie, 11 Rue des 4 vents", 30, "Bâtiment Normandie", "11 Rue des 4 vents"},
		{"Résidence Corneille Pav 72, 80 rue Corneille", 30, "Résidence Corneille Pav 72", "80 rue Corneille"},
		{"Résidence 3 - 65Ter, 4 Allée du bon lait", 16, "Résidence 3", "65Ter, 4 Allée du bon lait"},
		{"Résidence 3 - 65Ter, 4 Allée du bon lait", 10, "Résidence", "3 - 65Ter, 4 Allée du bon lait"},
		{"Résidence 3 - 65Ter, 4 Allée du bon lait", 11, "Résidence", "3 - 65Ter, 4 Allée du bon lait"},
		{"Une longue adresse sans tirets ni virgule", 20, "Une longue adresse", "sans tirets ni virgule"},
		{"Une_longue_adresse_sans tirets ni virgule ni espace", 20, "Une_longue_adresse_s", "ans tirets ni virgule ni espace"},
	}
	for _, tt := range tests {
		got, got2 := formatAdresse(tt.adresse, tt.maxWidth)
		tu.Assert(t, got == tt.want && got2 == tt.want2)
	}
}
