package sql

import (
	"testing"
	"time"

	tu "registro/utils/testutils"

	_ "github.com/lib/pq"
)

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

	// invalid currency
	_, err = db.Exec("UPDATE t1 SET montant = $1", Montant{0, 0xFF})
	tu.AssertNoErr(t, err)
	row := db.QueryRow("SELECT montant FROM t1;")
	err = row.Scan(&Montant{})
	tu.Assert(t, err != nil)
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

func TestDate(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TABLE t1 (id serial, date date);
	`)
	tu.AssertNoErr(t, err)

	ti := time.Date(2025, 5, 23, 12, 5, 6, 5, time.Local)
	_, err = db.Exec("INSERT INTO t1 (date) VALUES ($1)", ti)
	tu.AssertNoErr(t, err)

	var v time.Time
	err = db.QueryRow("SELECT date FROM t1;").Scan(&v)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v.Year() == ti.Year() && v.Month() == ti.Month() && v.Day() == ti.Day())
}
