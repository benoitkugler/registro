package logic

import (
	"testing"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	tu "registro/utils/testutils"
)

func TestDossierFinance_Bilan(t *testing.T) {
	taux := ds.Taux{Euros: 1000, FrancsSuisse: 2000}
	camps := cps.Camps{
		1: cps.Camp{Prix: ds.NewEuros(200), Duree: 10},
		2: cps.Camp{Prix: ds.NewFrancsuisses(150)},
	}
	df := DossierFinance{
		Dossier: Dossier{camps: camps}, taux: taux,
		paiements: ds.Paiements{
			1: ds.Paiement{IsRemboursement: true, Montant: ds.NewEuros(100)},
			2: ds.Paiement{IsRemboursement: false, Montant: ds.NewEuros(200)},
			3: ds.Paiement{IsRemboursement: false, Montant: ds.NewFrancsuisses(300)},
		},
	}

	// simple
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	tu.Assert(t, df.Bilan() == BilanFinances{ds.FrancsSuisse, 40000, 35000})

	// avec aides
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	df.aides = map[cps.IdParticipant]cps.Aides{
		1: {
			1: cps.Aide{Valide: true, Valeur: ds.NewEuros(20)},
			2: cps.Aide{Valide: true, Valeur: ds.NewFrancsuisses(2), ParJour: true},
		},
	}
	tu.Assert(t, df.Bilan() == BilanFinances{ds.FrancsSuisse, 40000 - 1000 - 200*10, 35000})

	// avec remises
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit, Remises: cps.Remises{
			ReducSpeciale:  ds.NewEuros(10),
			ReducEquipiers: 10,
			ReducEnfants:   5,
		}},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	df.aides = nil
	tu.Assert(t, df.Bilan() == BilanFinances{ds.FrancsSuisse, 40000 - 500 - 1500, 35000})

	// avec remises et aides
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit, Remises: cps.Remises{
			ReducSpeciale:  ds.NewEuros(10),
			ReducEquipiers: 5,
			ReducEnfants:   5,
		}},
	}
	df.aides = map[cps.IdParticipant]cps.Aides{
		1: {
			1: cps.Aide{Valide: true, Valeur: ds.NewEuros(20)},
		},
	}
	tu.Assert(t, df.Bilan() == BilanFinances{ds.FrancsSuisse, 10000 - 1000 - 500 - 900, 35000})
}
