package dossiers

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	"registro/sql/shared"
)

// RestrictByValidated only keep the [Dossier] with the given
// validation state.
func (dossiers Dossiers) RestrictByValidated(validated bool) {
	for id, dossier := range dossiers {
		if dossier.IsValidated != validated {
			delete(dossiers, id)
		}
	}
}

// Remise applique une remise de [percent]% sur le montant,
// le résultat étant pallié à 0 si besoin.
func (m Montant) Remise(percent int) Montant {
	p := m.Cent * (100 - percent) / 100
	if p < 0 {
		p = 0
	}
	return Montant{p, m.Currency}
}

func (m Montant) String() string {
	var val string
	if m.Cent%100 == 0 {
		val = strconv.Itoa(m.Cent / 100)
	} else {
		val = fmt.Sprintf("%.02f", float64(m.Cent)/100)
	}
	switch m.Currency {
	case FrancsSuisse:
		return m.Currency.String() + " " + val
	default:
		return val + m.Currency.String()
	}
}

// Has returns 'true' is the taux is able to handle the
// given [currency]
func (ts Taux) Has(currency Currency) bool {
	return newTableTaux(ts)[currency] != 0
}

// currency -> (1 currency = val / 1000 €)
type tableTaux [nbCurrencies]int

func newTableTaux(t Taux) tableTaux { return tableTaux{t.Euros, t.FrancsSuisse} }

// It will panic if one of the rate for [newCurrency] is 0
func (taux tableTaux) convertTo(origin Montant, newCurrency Currency) Montant {
	// Notations des taux : [origin.Currency] = Uo, [newCurrency] = Un
	// [origin] exprimé en [newCurrency] vaut donc : origin * Uo / Un
	Uo, Un := taux[origin.Currency], taux[newCurrency]
	converted := int(math.Round(float64(origin.Cent*Uo) / float64(Un)))
	return Montant{converted, newCurrency}
}

// MontantTaux ajoute une table de conversion
// à un montant, permettant :
//   - de l'afficher dans toutes les unités voulues
//   - des calculs entre différentes monnaies
type MontantTaux struct {
	Montant
	taux tableTaux
}

// Convertible is a shortcut for Zero() then Add(m)
func (t Taux) Convertible(m Montant) MontantTaux {
	out := t.Zero()
	out.Add(m)
	return out
}

// Zero return 0, expressed in the units with the higher taux.
// This is required to avoid conversion rounding errors.
func (t Taux) Zero() MontantTaux {
	table := newTableTaux(t)
	higherTaux := 0
	higherCurrency := Euros
	for currency, v := range table {
		if v > higherTaux {
			higherTaux = v
			higherCurrency = Currency(currency)
		}
	}
	return MontantTaux{Montant{Currency: higherCurrency}, newTableTaux(t)}
}

// Add ajoute [other], en convertissant correctement l'unité si besoin.
//
// La fonction 'panic' si le taux de [m.Montant.Currency] vaut 0.
func (m *MontantTaux) Add(other Montant) {
	m.Montant.Cent += m.taux.convertTo(other, m.Montant.Currency).Cent
}

// Sub soustrait [other], en convertissant correctement l'unité si besoin.
//
// La fonction 'panic' si le taux de [m.Montant.Currency] vaut 0.
func (m *MontantTaux) Sub(other Montant) {
	m.Montant.Cent -= m.taux.convertTo(other, m.Montant.Currency).Cent
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

// DescriptionHTML renvoie une description au format HTML
func (r Paiement) DescriptionHTML() template.HTML {
	var payeur string
	if r.IsRemboursement {
		payeur = fmt.Sprintf("Remboursement au %s", shared.NewDateFrom(r.Time))
	} else {
		payeur = fmt.Sprintf("Paiement de <i>%s</i> au %s", r.Payeur, shared.NewDateFrom(r.Time))
	}
	m := r.Montant
	if r.IsRemboursement {
		m.Cent *= -1
	}
	return template.HTML(payeur)
}
