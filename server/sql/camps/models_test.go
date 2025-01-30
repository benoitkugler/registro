package camps

import (
	"testing"
	"time"

	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	camp, err := randCamp().Insert(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, camp.Id != 0)
	personne, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp.Id, Roles: Roles{Direction, AutreRole}}.Insert(db)
	tu.AssertNoErr(t, err)
	// only one directeur
	_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp.Id, Roles: Roles{Direction, Menage}}.Insert(db)
	tu.Assert(t, err != nil)
}

func TestMontant(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TYPE montant AS (cent int, currenty smallint);

	CREATE TABLE t1 (id serial, montant montant);
	INSERT INTO t1 (id, montant) VALUES (1, (0, 0));
	`)
	tu.AssertNoErr(t, err)

	for _, expected := range [...]Montant{
		{0, 0},
		{-2, 1},
		{10, 2},
	} {
		_, err = db.Exec("UPDATE t1 SET montant = $1", expected)
		tu.AssertNoErr(t, err)

		var montant Montant
		row := db.QueryRow("SELECT montant FROM t1;")
		err = row.Scan(&montant)
		tu.AssertNoErr(t, err)
		tu.Assert(t, montant == expected)

	}
}

func TestMontant_String(t *testing.T) {
	for _, test := range [...]struct {
		m        Montant
		expected string
	}{
		{Montant{}, "0 <invalid currency>"},
		{Montant{0, 1}, "0 €"},
		{Montant{100, 1}, "1 €"},
		{Montant{-100, 1}, "-1 €"},
		{Montant{110, 1}, "1,1 €"},
		{Montant{110, 2}, "1,1 CHF"},
		{Montant{11589, 2}, "115,89 CHF"},
	} {
		tu.Assert(t, test.m.String() == test.expected)
	}
}

func TestCamp_DateFin(t *testing.T) {
	for _, test := range []struct {
		debut    shared.Date
		duree    int
		expected shared.Date
	}{
		{shared.NewDate(2000, time.January, 15), 1, shared.NewDate(2000, time.January, 15)},
		{shared.NewDate(2000, time.January, 15), 10, shared.NewDate(2000, time.January, 24)},
		{shared.NewDate(2000, time.January, 30), 3, shared.NewDate(2000, time.February, 1)},
		{shared.NewDate(2000, time.December, 30), 3, shared.NewDate(2001, time.January, 1)},
	} {
		got := (&Camp{DateDebut: test.debut, Duree: test.duree}).DateFin()
		tu.Assert(t, got == test.expected)
	}
}
