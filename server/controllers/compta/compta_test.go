package compta

import (
	"registro/sql/camps"
	ca "registro/sql/compta"
	"registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
	"testing"
	"time"
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

// func TestDetails(t *testing.T) {
// 	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
// 		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
// 		"../../migrations/init.sql")
// 	defer db.Remove()

// 	c1, c2 := createCamps(t, db)

// 	ct := Controller{db: db.DB}

// 	co, err := ct.loadOrCreateCompta(c1.Id)
// 	tu.AssertNoErr(t, err)

// 	details, err := ct.getCamp(co.Id)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, details.Prix == 0)
// 	tu.Assert(t, details.Inscrits == 0)
// 	tu.Assert(t, details.Equipiers == 1)
// 	tu.Assert(t, details.Animateurs == 1)
// 	tu.Assert(t, details.Duree == 8)
// 	tu.Assert(t, details.Directeur == " ")
// 	tu.Assert(t, details.Budget == 0)

// 	_, _, err = ct.getEcritures(co.Id)
// 	tu.AssertNoErr(t, err)

// 	co2, err := ct.loadOrCreateCompta(c2.Id)
// 	tu.AssertNoErr(t, err)

// 	_, _, err = ct.getBilanCamps(BilanCampsIn{Ids: []IdCompta{co.Id, co2.Id}})
// 	tu.AssertNoErr(t, err)
// }
