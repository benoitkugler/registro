package compta

import (
	"testing"

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

// func TestFraismdr(t *testing.T) {
// 	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
// 		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
// 		"../../migrations/init.sql")
// 	defer db.Remove()

// 	c1, _ := createCamps(t, db)

// 	ct := Controller{db: db.DB}

// 	co, err := ct.loadOrCreateCompta(c1.Id)
// 	tu.AssertNoErr(t, err)

// 	rec1, err := ct.createRecord(CreateRecordIn{co.Id, true})
// 	tu.AssertNoErr(t, err)
// 	rec2, err := ct.createRecord(CreateRecordIn{co.Id, false})
// 	tu.AssertNoErr(t, err)

// 	rec1.Records[0].Montants = Montants{{Kind: Carburant, Montant: 100}}
// 	_, err = ct.updateRecord(rec1, co.Id)
// 	tu.AssertNoErr(t, err)
// 	rec2.Records[0].Montants = Montants{{Kind: Carburant, Montant: 200}}
// 	_, err = ct.updateRecord(rec2, co.Id)
// 	tu.AssertNoErr(t, err)

// 	err = ct.setBusExpense(SetBusExpenseIn{IdCompta: co.Id, Records: []IdSubrecord{rec1.Records[0].Id}})
// 	tu.AssertNoErr(t, err)

// 	exp, err := ct.loadBusExpenses(co.Id)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, len(exp.Records) == 2 && len(exp.Selected) == 1)
// }
