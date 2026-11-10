package compta

import (
	"testing"

	ca "registro/sql/compta"
	tu "registro/utils/testutils"
)

func TestAtMDR(t *testing.T) {
	atMDRL := [...]string{
		"Chamaloc ",
		"Chamaloc",
		"maison du Rocher",
		"Maison du Rocher ",
		"Maison du Rocher",
		"Maison du rocher",
		"La Maison du Rocher",
	}
	notAtMDRL := [...]string{
		"Var",
		"Landes",
		"Chambon/Lignon",
		"St Julien St Alban",
		"Chambon sur Lignon -",
		"Chambon sur Lignon",
		"Bandol",
		"Hte Loire",
		"A définir Alboussière ou Champis",
		"Haute-Loire",
		"Six-four les plages",
		"Eglise quartier Moulin Champis",
		"Cévennes",
		"Champis",
		"Six-Fours-les-Plages",
		"landes",
		"Menditte",
		"Six-Four les plages",
		"Grenoble",
		"Zénica",
		"Champfleuri",
		"Corse",
		"Beauvezer (04)",
		"St Agnan en Vercors",
		"Valence",
		"St Julien",
		"La Cadière d'Azur",
		"Montmeyran",
		"Roumagnan",
		"Ardèche",
		"Arcelet",
	}
	for _, lieu := range atMDRL {
		tu.Assert(t, atMDR(lieu))
	}
	for _, lieu := range notAtMDRL {
		tu.Assert(t, !atMDR(lieu))
	}
}

func TestFraismdr_resolve(t *testing.T) {
	got := resolveFraismdr(ca.FraismdrACVE{
		Bus: []ca.Bus{
			{Name: "Car", Kind: ca.CarBus, KmStart: 0, KmEnd: 1000},
			{Name: "A", Kind: ca.CarMinibus, KmStart: 0, KmEnd: 1000},
			{Name: "A", Kind: ca.CarMinibus, KmStart: 0, KmEnd: 1000},
			{Name: "B", Kind: ca.CarCar, KmStart: 0, KmEnd: 1000},
		},
	}, ca.ComptaConfig{
		AssoCars:     ca.CarsCosts{20, 10, 5},
		EntretienMDR: 10,
		MaterielMDR:  5,
	}, nil, nil, 5, 20, 10, true)

	tu.Assert(t, len(got.Bus) == 4)
	tu.Assert(t, got.TotalBus == 1000*20+2*1000*10+1000*5)
	tu.Assert(t, got.Materiel.CostEntretien == 5*30*10)
	tu.Assert(t, got.Materiel.CostMateriel == 5*20*5)

	got = resolveFraismdr(ca.FraismdrACVE{
		Bus: []ca.Bus{
			{Name: "Car", Kind: ca.CarBus, KmStart: 0, KmEnd: 1000},
			{Name: "A", Kind: ca.CarMinibus, KmStart: 0, KmEnd: 1000},
			{Name: "A", Kind: ca.CarMinibus, KmStart: 0, KmEnd: 1000},
			{Name: "B", Kind: ca.CarCar, KmStart: 0, KmEnd: 1000},
		},
	}, ca.ComptaConfig{
		AssoCars:     ca.CarsCosts{20, 10, 5},
		EntretienMDR: 10,
		MaterielMDR:  5,
	}, ca.ComptaSubrecords{
		1: {Montants: ca.Montants{{Kind: ca.Travels, Cent: 100}, {Kind: ca.Alimentation, Cent: 200}}},
		2: {Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 200}}},
		3: {Montants: ca.Montants{{Kind: ca.Carburant, Cent: 300}, {Kind: ca.Alimentation, Cent: 200}}},
	}, ca.ComptaRecordBuss{{IdSubrecord: 1}, {IdSubrecord: 2}}, 5, 20, 10, true)

	tu.Assert(t, got.TotalBus == 1000*20+2*1000*10+1000*5-300)
}

// TODO:
// func TestFraismdr(t *testing.T) {
// 	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
// 		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
// 		"../../migrations/init.sql")
// 	defer db.Remove()

// 	c1, _ := createCamps(t, db)

// 	ct := Controller{db: db.DB}

// 	co, err := ct.loadOrCreateCompta(c1.Id)
// 	tu.AssertNoErr(t, err)

// 	rec1, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountAsso})
// 	tu.AssertNoErr(t, err)
// 	rec2, err := ct.createRecord(CreateRecordIn{co.Id, ca.AccountCamp})
// 	tu.AssertNoErr(t, err)

// 	rec1.Subrecords[0].Montants = ca.Montants{{Kind: ca.Carburant, Cent: 100}}
// 	_, err = ct.updateRecord(rec1, co.Id)
// 	tu.AssertNoErr(t, err)
// 	rec2.Subrecords[0].Montants = ca.Montants{{Kind: ca.Carburant, Cent: 200}}
// 	_, err = ct.updateRecord(rec2, co.Id)
// 	tu.AssertNoErr(t, err)

// 	err = ct.setBusExpense(SetBusExpenseIn{IdCompta: co.Id, Records: []IdSubrecord{rec1.Records[0].Id}})
// 	tu.AssertNoErr(t, err)

// 	exp, err := ct.loadBusExpenses(co.Id)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, len(exp.Records) == 2 && len(exp.Selected) == 1)
// }
