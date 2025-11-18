package logic

import (
	"reflect"
	"testing"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	tu "registro/utils/testutils"
)

var (
	eur = ds.NewEuros
	chf = ds.NewFrancsuisses
)

func TestDossierFinance_Bilan(t *testing.T) {
	taux := ds.Taux{Euros: 1000, FrancsSuisse: 2000}
	camps := cps.Camps{
		1: cps.Camp{Prix: eur(200), Duree: 10},
		2: cps.Camp{Prix: chf(150)},
	}
	df := DossierFinance{
		Dossier: Dossier{camps: camps}, Taux: taux,
		paiements: ds.Paiements{
			1: ds.Paiement{IsRemboursement: true, Montant: eur(100)},
			2: ds.Paiement{IsRemboursement: false, Montant: eur(200)},
			3: ds.Paiement{IsRemboursement: false, Montant: chf(300)},
		},
	}

	// simple
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), "", cps.Remises{}, nil},
			2: {chf(150), "", cps.Remises{}, nil},
			3: {chf(150), "", cps.Remises{}, nil},
		},
		40000, 15000, 35000, 0, 0, ds.FrancsSuisse,
	}))

	// avec aides
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	df.aides = map[cps.IdParticipant]cps.Aides{
		1: {
			1: cps.Aide{Valide: true, Valeur: eur(20)},
			2: cps.Aide{Valide: true, Valeur: chf(2), ParJour: true},
			3: cps.Aide{Valide: false, Valeur: chf(5)},
		},
	}
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), "", cps.Remises{}, []AideResolved{{"", eur(20)}, {"", chf(20)}}},
			2: {chf(150), "", cps.Remises{}, nil},
			3: {chf(150), "", cps.Remises{}, nil},
		},
		40000 - 1000 - 200*10, 15000, 35000, 0, 1000 + 200*10, ds.FrancsSuisse,
	}))

	// avec remises
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit, Remises: cps.Remises{
			Speciale:  eur(10),
			Equipiers: 10,
			Famille:   5,
		}},
		2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
		3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
		4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
	}
	df.aides = nil
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), "", cps.Remises{
				Speciale:  eur(10),
				Equipiers: 10,
				Famille:   5,
			}, nil},
			2: {chf(150), "", cps.Remises{}, nil},
			3: {chf(150), "", cps.Remises{}, nil},
		},
		40000 - 500 - 1500, 15000, 35000, 0, 0, ds.FrancsSuisse,
	}))

	// avec remises et aides
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit, Remises: cps.Remises{
			Speciale:  eur(10),
			Equipiers: 5,
			Famille:   5,
		}},
	}
	df.aides = map[cps.IdParticipant]cps.Aides{
		1: {
			1: cps.Aide{Valide: true, Valeur: eur(20)},
		},
	}
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), "", cps.Remises{
				Speciale:  eur(10),
				Equipiers: 5,
				Famille:   5,
			}, []AideResolved{{"", eur(20)}}},
		},
		10000 - 1000 - 500 - 900, 0, 35000, 0, 1000, ds.FrancsSuisse,
	}))

	// avec fond soutien
	df = DossierFinance{
		Dossier: Dossier{camps: camps}, Taux: taux,
		paiements: ds.Paiements{
			1: ds.Paiement{IsRemboursement: true, Montant: eur(100), Payeur: ds.PayeurFondSoutien},
			2: ds.Paiement{IsRemboursement: false, Montant: eur(200), Payeur: ds.PayeurFondSoutien},
			3: ds.Paiement{IsRemboursement: false, Montant: chf(300)},
		},
	}
	df.Participants = cps.Participants{
		1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
	}
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), "", cps.Remises{}, nil},
		},
		10000, 0, 35000, 5000, 0, ds.FrancsSuisse,
	}))
}

func Test_pc_prixBase(t *testing.T) {
	status := []cps.PrixParStatut{{Id: 1, Prix: 8000, Label: "Enfant", Description: ""}, {Id: 2, Prix: 9000, Label: "Adulte", Description: ""}}
	jours := []int{1000, 2000, 3000, 4000}
	type fields struct {
		optPart cps.OptionPrixParticipant
		optCamp cps.OptionPrixCamp
		optQF   cps.PrixQuotientFamilial
		prix    cps.Montant
		duree   int
		qf      int
	}
	tests := []struct {
		fields fields
		want   cps.Montant
		want1  string
	}{
		// Simple
		{fields{prix: eur(100)}, eur(100), ""},
		// QF
		{fields{prix: eur(100), qf: 100}, eur(100), ""},
		{fields{prix: eur(100), optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}}, eur(100), ""},
		{fields{prix: eur(100), optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 100}, eur(20), "QF 100"},
		{fields{prix: eur(100), optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 400}, eur(40), "QF 400"},
		{fields{prix: eur(100), optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 600}, eur(60), "QF 600"},
		{fields{prix: eur(100), optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 1000}, eur(100), "QF 1000"},
		// Option statut
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixStatut, Statuts: status}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixStatut, Statuts: status}, optPart: cps.OptionPrixParticipant{IdStatut: 1}}, eur(80), "Enfant"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixStatut, Statuts: status}, optPart: cps.OptionPrixParticipant{IdStatut: 2}}, eur(90), "Adulte"},
		// Option jours
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, Jours: jours}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, Jours: jours}, optPart: cps.OptionPrixParticipant{Jour: cps.Jours{}}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, Jours: jours}, optPart: cps.OptionPrixParticipant{Jour: cps.Jours{0, 1, 2, 3}}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, Jours: jours}, optPart: cps.OptionPrixParticipant{Jour: cps.Jours{0, 2}}}, eur(40), "2 jours"},
		// Mélange
		{
			fields{
				prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, Jours: jours},
				optPart: cps.OptionPrixParticipant{Jour: cps.Jours{0, 2}},
				optQF:   cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 100,
			}, eur(8), "2 jours - QF 100",
		}, // 40 * 20%
	}
	for _, tt := range tests {
		p := pc{
			Participant: cps.Participant{OptionPrix: tt.fields.optPart, QuotientFamilial: tt.fields.qf},
			Camp: cps.Camp{
				OptionPrix:             tt.fields.optCamp,
				OptionQuotientFamilial: tt.fields.optQF,
				Prix:                   tt.fields.prix, Duree: tt.fields.duree,
			},
		}
		got, got1 := p.prixBase()
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got1 == tt.want1)
	}
}
