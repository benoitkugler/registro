package dossiers

import (
	"math/rand"
	"testing"

	tu "registro/utils/testutils"
)

func TestMontantTaux_Add(t *testing.T) {
	tests := []struct {
		taux  tableTaux
		val   Montant
		other Montant
		want  Montant
	}{
		// trivial cases
		{tableTaux{1000, 1000}, Montant{100, 0}, Montant{200, 0}, Montant{300, 0}},
		{tableTaux{1000, 1000}, Montant{100, 0}, Montant{200, 1}, Montant{300, 0}},
		{tableTaux{1000, 2000}, Montant{100, 0}, Montant{200, 0}, Montant{300, 0}},
		{tableTaux{1000, 2000}, Montant{100, 1}, Montant{200, 1}, Montant{300, 1}},
		// real conversion : 1CHF = 2€
		{tableTaux{1000, 2000}, Montant{100, 0}, Montant{200, 1}, Montant{500, 0}},
		{tableTaux{1000, 2000}, Montant{100, 1}, Montant{200, 0}, Montant{200, 1}},
		{tableTaux{1000, 2000}, Montant{100, 1}, Montant{200, 1}, Montant{300, 1}},
		// real conversion : 1CHF = 2€, avec virgules
		{tableTaux{1000, 2000}, Montant{155, 0}, Montant{200, 1}, Montant{555, 0}},
		{tableTaux{1000, 2000}, Montant{123, 1}, Montant{202, 0}, Montant{224, 1}},
		{tableTaux{1000, 2000}, Montant{100, 1}, Montant{201, 0}, Montant{201, 1}},
	}
	for _, tt := range tests {
		m := &MontantTaux{
			Montant: tt.val,
			taux:    tt.taux,
		}
		m.Add(tt.other)
		tu.Assert(t, m.Montant == tt.want)
	}
}

func TestMontant_String(t *testing.T) {
	for _, test := range [...]struct {
		m        Montant
		expected string
	}{
		{Montant{0, Euros}, "0€"},
		{Montant{100, Euros}, "1€"},
		{Montant{-100, Euros}, "-1€"},
		{Montant{110, Euros}, "1,1€"},
		{Montant{110, FrancsSuisse}, "CHF 1,1"},
		{Montant{11589, FrancsSuisse}, "CHF 115,89"},
		{Montant{0, nbCurrencies}, "0<invalid currency>"},
	} {
		tu.Assert(t, test.m.String() == test.expected)
	}
}

func TestMontantTaux_String(t *testing.T) {
	tests := []struct {
		taux    tableTaux
		Montant Montant
		want    string
	}{
		{tableTaux{}, Montant{}, ""},
		{tableTaux{1000}, Montant{100, Euros}, "1€"},
		{tableTaux{0, 1000}, Montant{100, FrancsSuisse}, "CHF 1"},
		// 1CHF = 2€
		{tableTaux{1000, 2000}, Montant{100, Euros}, "1€ ou CHF 0,5"},
		{tableTaux{1000, 2000}, Montant{182, Euros}, "1,82€ ou CHF 0,91"},
		{tableTaux{1000, 2000}, Montant{100, FrancsSuisse}, "2€ ou CHF 1"},
	}
	for _, tt := range tests {
		m := &MontantTaux{
			Montant: tt.Montant,
			taux:    tt.taux,
		}
		tu.Assert(t, m.String() == tt.want)
	}
}

func TestConversionsRoundtrip(t *testing.T) {
	tu.Assert(t, Taux{Euros: 1000}.Zero().Currency == Euros)
	tu.Assert(t, Taux{Euros: 1000, FrancsSuisse: 800}.Zero().Currency == Euros)
	tu.Assert(t, Taux{Euros: 1000, FrancsSuisse: 1106}.Zero().Currency == FrancsSuisse)

	for _, tauxCHF := range []int{
		1000, 1100, 1110, 1111, 1106, 1007, 1230,
		900, 850, 999, 990, 756,
	} {
		taux := Taux{Euros: 1000, FrancsSuisse: tauxCHF}
		table := taux.Zero().taux
		for range [20000]bool{} {
			prix := Montant{rand.Intn(200000), Euros}
			prixC := taux.Zero()
			prixC.Add(prix)

			// paiement in either units
			inCHF := table.convertTo(prix, FrancsSuisse)
			inEuros := table.convertTo(prix, Euros)

			paiementCHF := taux.Zero()
			paiementCHF.Add(inCHF)

			paiementEuros := taux.Zero()
			paiementEuros.Add(inEuros)

			// check that converting back does not induce rouding errors
			tu.Assert(t, prixC.Montant == paiementCHF.Montant)
			tu.Assert(t, prixC.Montant == paiementEuros.Montant)
		}
	}
}

func TestMontant_Remise(t *testing.T) {
	tests := []struct {
		Cent    int
		percent int
		want    int
	}{
		{100, 50, 50},
		{100, 0, 100},
		{200, 0, 200},
		{200, 100, 0},
		{200, 25, 150},
		{200, 150, 0},
		{200, 33, 200 - 66},
	}
	for _, tt := range tests {
		m := Montant{Cent: tt.Cent}
		tu.Assert(t, m.Remise(tt.percent) == Montant{Cent: tt.want})
	}
}
