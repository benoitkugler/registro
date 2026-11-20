package logic

import (
	"reflect"
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	sh "registro/sql/shared"
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
	bilan := df.Bilan()
	tu.Assert(t, reflect.DeepEqual(bilan, BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), eur(200), "", cps.Remises{}, nil},
			2: {chf(150), chf(150), "", cps.Remises{}, nil},
			3: {chf(150), chf(150), "", cps.Remises{}, nil},
		},
		40000, 15000, 35000, 0, 0, ds.FrancsSuisse,
	}))
	tu.Assert(t, bilan.Inscrits[1].ResolvedRemises(taux) == chf(0))

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
	bilan = df.Bilan()
	tu.Assert(t, reflect.DeepEqual(bilan, BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), eur(200), "", cps.Remises{}, []AideResolved{{"", eur(20)}, {"", chf(20)}}},
			2: {chf(150), chf(150), "", cps.Remises{}, nil},
			3: {chf(150), chf(150), "", cps.Remises{}, nil},
		},
		40000 - 1000 - 200*10, 15000, 35000, 0, 1000 + 200*10, ds.FrancsSuisse,
	}))
	tu.Assert(t, bilan.Inscrits[1].ResolvedRemises(taux) == chf(0))

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
	bilan = df.Bilan()
	tu.Assert(t, reflect.DeepEqual(bilan, BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), eur(200), "", cps.Remises{
				Speciale:  eur(10),
				Equipiers: 10,
				Famille:   5,
			}, nil},
			2: {chf(150), chf(150), "", cps.Remises{}, nil},
			3: {chf(150), chf(150), "", cps.Remises{}, nil},
		},
		40000 - 500 - 1500, 15000, 35000, 0, 0, ds.FrancsSuisse,
	}))
	tu.Assert(t, bilan.Inscrits[1].ResolvedRemises(taux) == chf(20))

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
	bilan = df.Bilan()
	tu.Assert(t, reflect.DeepEqual(df.Bilan(), BilanFinances{
		map[cps.IdParticipant]BilanParticipant{
			1: {eur(200), eur(200), "", cps.Remises{
				Speciale:  eur(10),
				Equipiers: 5,
				Famille:   5,
			}, []AideResolved{{"", eur(20)}}},
		},
		10000 - 1000 - 500 - 900, 0, 35000, 0, 1000, ds.FrancsSuisse,
	}))
	tu.Assert(t, bilan.Inscrits[1].ResolvedRemises(taux) == chf(14))

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
			1: {eur(200), eur(200), "", cps.Remises{}, nil},
		},
		10000, 0, 35000, 5000, 0, ds.FrancsSuisse,
	}))
}

func TestDossierFinance_MontantsPayesParDevise(t *testing.T) {
	df := DossierFinance{paiements: ds.Paiements{
		1: {Montant: eur(100)},
		2: {Montant: chf(80), Payeur: ds.PayeurFondSoutien},
		3: {Montant: eur(25), IsRemboursement: true},
		4: {Montant: chf(10), IsRemboursement: true},
	}}

	got := df.MontantsPayesParDevise()
	tu.Assert(t, got[ds.Euros] == 7500)
	tu.Assert(t, got[ds.FrancsSuisse] == 7000)
}

func TestDossierFinance_DispatchPaiements(t *testing.T) {
	tests := []struct {
		name string
		df   DossierFinance
		want map[cps.IdParticipant]ds.Montant
	}{
		{
			name: "taux EUR/CHF, paiements EUR et CHF",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{
						1: cps.Camp{Prix: eur(200), Duree: 10},
						2: cps.Camp{Prix: chf(150)},
					},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
						2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
						3: cps.Participant{Id: 3, IdCamp: 2, Statut: cps.Inscrit},
						4: cps.Participant{Id: 4, IdCamp: 2, Statut: cps.AStatuer},
					},
				},
				Taux: ds.Taux{Euros: 1000, FrancsSuisse: 2000},
				paiements: ds.Paiements{
					1: {Montant: eur(100), IsRemboursement: true},
					2: {Montant: eur(200)},
					3: {Montant: chf(300)},
				},
			},
			want: map[cps.IdParticipant]ds.Montant{1: chf(87.50), 2: chf(131.25), 3: chf(131.25)},
		},
		{
			name: "taux EUR seulement, paiements EUR",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{
						1: cps.Camp{Prix: eur(200)},
						2: cps.Camp{Prix: eur(100)},
					},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
						2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
					},
				},
				Taux: ds.Taux{Euros: 1000},
				paiements: ds.Paiements{
					1: {Montant: eur(200)},
					2: {Montant: eur(50), IsRemboursement: true},
				},
			},
			want: map[cps.IdParticipant]ds.Montant{1: eur(100.00), 2: eur(50.00)},
		},
		{
			name: "taux EUR/CHF, paiements CHF",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{
						1: cps.Camp{Prix: eur(100)},
						2: cps.Camp{Prix: eur(300)},
					},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
						2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
					},
				},
				Taux: ds.Taux{Euros: 1000, FrancsSuisse: 2000},
				paiements: ds.Paiements{
					1: {Montant: chf(200)},
					2: {Montant: chf(50)},
				},
			},
			want: map[cps.IdParticipant]ds.Montant{1: chf(62.50), 2: chf(187.50)},
		},
		{
			name: "taux EUR/CHF, paiements EUR",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{
						1: cps.Camp{Prix: chf(100)},
						2: cps.Camp{Prix: chf(300)},
					},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
						2: cps.Participant{Id: 2, IdCamp: 2, Statut: cps.Inscrit},
					},
				},
				Taux: ds.Taux{Euros: 1000, FrancsSuisse: 2000},
				paiements: ds.Paiements{
					1: {Montant: eur(150)},
					2: {Montant: eur(50)},
				},
			},
			want: map[cps.IdParticipant]ds.Montant{1: chf(25.00), 2: chf(75.00)},
		},
		{
			name: "aucun inscrit",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{1: cps.Camp{Prix: chf(100)}},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.AStatuer},
					},
				},
				Taux: ds.Taux{Euros: 1000, FrancsSuisse: 2000},
				paiements: ds.Paiements{
					1: {Montant: chf(50)},
				},
			},
			want: nil,
		},
		{
			name: "avec reste",
			df: DossierFinance{
				Dossier: Dossier{
					camps: cps.Camps{1: cps.Camp{Prix: chf(100)}},
					Participants: cps.Participants{
						1: cps.Participant{Id: 1, IdCamp: 1, Statut: cps.Inscrit},
						2: cps.Participant{Id: 2, IdCamp: 1, Statut: cps.Inscrit},
						3: cps.Participant{Id: 3, IdCamp: 1, Statut: cps.Inscrit},
					},
				},
				Taux: ds.Taux{Euros: 1000, FrancsSuisse: 2000},
				paiements: ds.Paiements{
					1: {Montant: eur(200)},
					2: {Montant: chf(400)},
				},
			},
			want: map[cps.IdParticipant]ds.Montant{1: chf(166.66), 2: chf(166.66), 3: chf(166.68)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.df.Bilan().DispatchPaiements(tt.df.Taux)
			tu.Assert(t, reflect.DeepEqual(got, tt.want))
		})
	}
}

func Test_pc_prixBase(t *testing.T) {
	inscription := time.Now()
	debut := sh.NewDate(2026, time.July, 1)
	gridAge := cps.PrixParAge{{AgeMax: 10, Prix: 5000}, {AgeMax: 18, Prix: 8000}}
	type fields struct {
		optPart       cps.OptionPrixParticipant
		optCamp       cps.OptionPrixCamp
		optQF         cps.PrixQuotientFamilial
		prix          cps.Montant
		duree         int
		qf            int
		dateNaissance sh.Date
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
		// Option rapide
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixInscriptionRapide, InscriptionRapide: cps.InscriptionRapide{Prix: 8000, Limite: sh.Date(inscription.Add(-5 * 24 * time.Hour))}}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixInscriptionRapide, InscriptionRapide: cps.InscriptionRapide{Prix: 8000, Limite: sh.Date(inscription)}}}, eur(80), "Inscription rapide"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixInscriptionRapide, InscriptionRapide: cps.InscriptionRapide{Prix: 8000, Limite: sh.Date(inscription.Add(5 * 24 * time.Hour))}}}, eur(80), "Inscription rapide"},
		// Option jours
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: 2000}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: 2000}, optPart: cps.OptionPrixParticipant{Jour: cps.JoursPresence{}}}, eur(100), ""},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: 2000}, optPart: cps.OptionPrixParticipant{Jour: cps.JoursPresence{0, 1, 2, 3}}}, eur(80), "4 jours"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: 2000}, optPart: cps.OptionPrixParticipant{Jour: cps.JoursPresence{0, 2}}}, eur(40), "2 jours"},
		// Option prix par âge
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: gridAge}, dateNaissance: sh.NewDate(2020, time.July, 1)}, eur(50), "Âge 6 ans"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: gridAge}, dateNaissance: sh.NewDate(2021, time.July, 1)}, eur(50), "Âge 5 ans"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: gridAge}, dateNaissance: sh.NewDate(2016, time.July, 1)}, eur(80), "Âge 10 ans"},
		{fields{prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: gridAge}, dateNaissance: sh.NewDate(2000, time.July, 1)}, eur(100), ""},
		// Mélange
		{
			fields{
				prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: 2000},
				optPart: cps.OptionPrixParticipant{Jour: cps.JoursPresence{0, 2}},
				optQF:   cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 100,
			}, eur(8), "2 jours - QF 100",
		}, // 40 * 20%
		{
			fields{
				prix: eur(100), optCamp: cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: gridAge},
				optQF: cps.PrixQuotientFamilial{20, 40, 60, 100}, qf: 100,
				dateNaissance: sh.NewDate(2020, time.July, 1),
			}, eur(10), "Âge 6 ans - QF 100",
		}, // 50 * 20%
	}
	for _, tt := range tests {
		p := partExt{
			Participant: cps.Participant{OptionPrix: tt.fields.optPart, QuotientFamilial: tt.fields.qf},
			Camp: cps.Camp{
				OptionPrix:             tt.fields.optCamp,
				OptionQuotientFamilial: tt.fields.optQF,
				Prix:                   tt.fields.prix, Duree: tt.fields.duree,
				DateDebut: debut,
			},
			dateNaissance: tt.fields.dateNaissance,
			inscription:   inscription,
		}
		got, got1 := p.prixBase()
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got1 == tt.want1)
	}
}
