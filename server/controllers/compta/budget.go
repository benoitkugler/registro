package compta

import ds "registro/sql/dossiers"

type BudgetPrevisionnelSummary struct {
	Reel   ds.Montant
	Simule ds.Montant
}
