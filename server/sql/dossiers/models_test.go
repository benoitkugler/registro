package dossiers

import (
	"testing"

	tu "registro/utils/testutils"
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
		{0, nbCurrencies},
		{-2, Euros},
		{10, FrancsSuisse},
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
