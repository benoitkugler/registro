package compta

import (
	"testing"
	"time"

	"registro/sql/camps"
	ca "registro/sql/compta"
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
