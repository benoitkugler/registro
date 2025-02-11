package dossiers

import (
	"fmt"
	"strings"

	"registro/sql/shared"
)

type OptIdTaux shared.OptID[IdTaux]

func (Id IdTaux) Opt() OptIdTaux { return OptIdTaux{Id: Id, Valid: true} }

// Mode de paiement
type ModePaiement uint8

const (
	Cheque  ModePaiement = iota
	EnLigne              // (carte bancaire, en ligne)
	Virement
	Especes
	Ancv
	// uniquement pour les dons
	Helloasso
)

// Montant représente un prix (avec son unité).
type Montant struct {
	Cent     int
	Currency Currency
}

func NewEuros(f float32) Montant { return Montant{int(f * 100), Euros} }

func (m Montant) String() string {
	val := strings.ReplaceAll(fmt.Sprintf("%g", float64(m.Cent)/100), ".", ",")
	switch m.Currency {
	case FrancsSuisse:
		return m.Currency.String() + " " + val
	default:
		return val + m.Currency.String()
	}
}

// Add assume the currency is the same.
func (s *Montant) Add(other Montant) { s.Cent += other.Cent }

type Currency uint8

const (
	Euros        Currency = iota // €
	FrancsSuisse                 // CHF
)

const nbCurrencies = FrancsSuisse + 1 // gomacro:no-enum

func (c Currency) String() string {
	switch c {
	case Euros:
		return "€"
	case FrancsSuisse:
		return "CHF"
	default:
		return "<invalid currency>"
	}
}

// currency -> (1 currency = val / 1000 €)
type tableTaux [nbCurrencies]int

// It will panic if one of the rate for [newCurrency] is 0
func (taux tableTaux) convertTo(origin Montant, newCurrency Currency) Montant {
	// Notations des taux : [origin.Currency] = Uo, [newCurrency] = Un
	// [origin] exprimé en [newCurrency] vaut donc : origin * Uo / Un
	Uo, Un := taux[origin.Currency], taux[newCurrency]
	return Montant{origin.Cent * Uo / Un, newCurrency}
}

// MontantTaux ajoute une table de conversion
// à un montant, permettant :
//   - de l'afficher dans toutes les unités voulues
//   - des calculs entre différentes monnaies
type MontantTaux struct {
	Montant
	taux tableTaux
}

func (t Taux) Convert(m Montant) MontantTaux {
	return MontantTaux{m, tableTaux{t.Euros, t.FrancsSuisse}}
}

// Add ajoute [other], en convertissant correctement l'unité si besoin.
//
// La fonction 'panic' si le taux de [m.Montant.Currency] vaut 0.
func (m *MontantTaux) Add(other Montant) {
	m.Montant.Cent += m.taux.convertTo(other, m.Montant.Currency).Cent
}

// String affiche le montant dans les unités pour
// lesquelles le taux n'est pas 0.
func (m MontantTaux) String() string {
	var chunks []string
	for currency, taux := range m.taux {
		if taux == 0 {
			continue
		}
		c := Currency(currency)
		chunks = append(chunks, m.taux.convertTo(m.Montant, c).String())
	}
	return strings.Join(chunks, " ou ")
}
