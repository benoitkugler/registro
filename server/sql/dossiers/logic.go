package dossiers

import "fmt"

// RestrictByValidated only keep the [Dossier] with the given
// validation state.
func (dossiers Dossiers) RestrictByValidated(validated bool) {
	for id, dossier := range dossiers {
		if dossier.IsValidated != validated {
			delete(dossiers, id)
		}
	}
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
	montant := fmt.Sprintf("<i>%s</i>", taux.Convert(r.Montant).String())
	return payeur, montant
}
