package dossiers

import (
	"testing"
	"time"

	"registro/sql/personnes"
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

func TestEvents(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	_, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = Dossier{IdTaux: 1, IdResponsable: 1}.Insert(db)

	event, err := Event{IdDossier: 1, Kind: Message, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventMessage{IdEvent: event.Id, Guard: Message}.Insert(db)
	tu.AssertNoErr(t, err)
}
