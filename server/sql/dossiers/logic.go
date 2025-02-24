package dossiers

import (
	"fmt"
	"math"
	"strings"
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

func (t Taux) Convertible(m Montant) MontantTaux { return MontantTaux{m, newTableTaux(t)} }

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
	return t.Convertible(Montant{Currency: higherCurrency})
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

// DescriptionHTML renvoie une description et le montant, au format HTML
func (r Paiement) DescriptionHTML(taux Taux) (string, string) {
	var payeur string
	if r.IsAcompte {
		payeur = fmt.Sprintf("Acompte de <i>%s</i> au %s", r.Payeur, r.Date)
	} else if r.IsRemboursement {
		payeur = fmt.Sprintf("Remboursement au %s", r.Date)
	} else {
		payeur = fmt.Sprintf("Paiement de <i>%s</i> au %s", r.Payeur, r.Date)
	}
	m := r.Montant
	if r.IsRemboursement {
		m.Cent *= -1
	}
	montant := fmt.Sprintf("<i>%s</i>", taux.Convertible(r.Montant).String())
	return payeur, montant
}
