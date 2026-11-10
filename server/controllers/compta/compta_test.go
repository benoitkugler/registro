package compta

import (
	"testing"
	"time"

	"registro/sql/camps"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	"registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func createCamps(t *testing.T, db ca.DB) (c1, c2 camps.Camp) {
	c1, err := camps.Camp{
		IdTaux:    1,
		Lieu:      "Chamaloc",
		DateDebut: shared.NewDateFrom(time.Now()),
		Duree:     7,
	}.Insert(db)
	tu.AssertNoErr(t, err)

	c2, err = camps.Camp{
		IdTaux:    1,
		Lieu:      "Var",
		DateDebut: shared.NewDateFrom(time.Now()),
		Duree:     10,
	}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = ca.ComptaConfig{Year: 2000}.Insert(db)
	tu.AssertNoErr(t, err)

	p, err := personnes.Personne{Identite: personnes.Identite{Mail: "bench26@gmail.com"}}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = camps.Equipier{IdPersonne: p.Id, IdCamp: c1.Id, Roles: camps.Roles{camps.Direction}}.Insert(db)
	tu.AssertNoErr(t, err)

	return
}

func TestState(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	_, _, smtp := devEnv(t)
	ct := Controller{db: db.DB, smtp: smtp, adminNotificationMail: "bench26@gmail.com"}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.setComptaState(SetComptaStateIn{IdCompta: co.Id, State: ca.ToBeValidated, SendMail: true}, userAdmin)
	tu.AssertNoErr(t, err)
}

func TestDetails(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, c2 := createCamps(t, db)

	ct := Controller{db: db.DB}

	co, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	details, err := ct.compileDetails(co.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, details.Prix.Cent == 0)
	tu.Assert(t, details.Inscrits == 0)
	tu.Assert(t, details.Equipiers == 1)
	tu.Assert(t, details.AuPairs == 1)
	tu.Assert(t, details.Duree == 7)
	tu.Assert(t, details.Directeur == " ")
	tu.Assert(t, details.Recettes.Cent == 0)

	_, _, err = ct.getEcritures(co.Id)
	tu.AssertNoErr(t, err)

	co2, err := ct.loadOrCreateCompta(c2.Id)
	tu.AssertNoErr(t, err)

	_, _, err = ct.getBilanCamps(BilanCampsIn{Ids: []ca.IdComptaCamp{co.Id, co2.Id}})
	tu.AssertNoErr(t, err)
}

func TestAggregateDepenses(t *testing.T) {
	ld := comptaLoader{
		records: ca.ComptaRecords{
			1: {Account: ca.AccountAsso, IsReceipe: true},
			2: {Account: ca.AccountAsso, IsReceipe: false},
			3: {Account: ca.AccountCamp, IsReceipe: true},
			4: {Account: ca.AccountCamp, IsReceipe: false},
		},
		subrecords: ca.ComptaSubrecords{
			1: {
				IdRecord: 1,
				Montants: ca.Montants{{Kind: ca.Travels, Cent: 100}, {Kind: ca.Alimentation, Cent: 200}, {Kind: ca.Alimentation, Cent: 200}},
			},
			2: {
				IdRecord: 1,
				Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 200}},
			},
			3: {
				IdRecord: 3,
				Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 250}},
			},
			4: {
				IdRecord: 2,
				Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 250}},
			},
			5: {
				IdRecord: 2,
				Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 250}},
			},
			6: {
				IdRecord: 4,
				Montants: ca.Montants{{Kind: ca.Carburant, Cent: 350}, {Kind: ca.Alimentation, Cent: 300}},
			},
		},
	}

	multi := func(cents int) ds.MultiCurrencies {
		return ds.Montant{Cent: cents, Currency: ds.Euros}.Multi()
	}
	fraisMDR := ResolvedFraisMDR{}
	expected := ventilations{
		ca.Travels: depenseTotale{accAssoReceipes: multi(100)},
		ca.Alimentation: {
			accAssoDepenses: multi(500), accAssoReceipes: multi(600),
			accCampDepenses: multi(300), accCampReceipes: multi(250),
		},
		ca.Carburant: {
			accAssoDepenses: multi(600), accAssoReceipes: multi(300),
			accCampDepenses: multi(350), accCampReceipes: multi(300),
		},
	}
	got := ld.aggregateDepenses(fraisMDR)
	tu.Assert(t, got == expected)
	campNet, assoNet := got.sum().net()
	taux := ds.Taux{Euros: 1000}
	tu.Assert(t, campNet.Agg(taux).Cent == 650-550)
	tu.Assert(t, assoNet.Agg(taux).Cent == 100)

	fraisMDR = ResolvedFraisMDR{
		HasMateriel: true,
		Materiel: ResolvedMateriel{
			CostEntretien: 100,
			CostMateriel:  200,
		},
		TotalBus: 300,
	}
	expected = ventilations{
		ca.Travels: depenseTotale{accAssoReceipes: multi(100)},
		ca.Alimentation: {
			accAssoDepenses: multi(500), accAssoReceipes: multi(600),
			accCampDepenses: multi(300), accCampReceipes: multi(250),
		},
		ca.Carburant: {
			accAssoDepenses: multi(900), accAssoReceipes: multi(300),
			accCampDepenses: multi(350), accCampReceipes: multi(300),
		},
		ca.FournitureEntretien:   depenseTotale{accAssoDepenses: multi(100)},
		ca.FourniturePedagogique: depenseTotale{accAssoDepenses: multi(200)},
	}
	tu.Assert(t, ld.aggregateDepenses(fraisMDR) == expected)
}
