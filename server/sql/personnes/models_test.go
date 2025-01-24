package personnes

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "gen_create.sql")
	defer db.Remove()

	p := randPersonne()
	p, err := p.Insert(db)
	tu.AssertNoErr(t, err)

	date := NewDate(2025, time.September, 23)
	p.DateNaissance = date
	_, err = p.Update(db)
	tu.AssertNoErr(t, err)

	p, err = SelectPersonne(db, p.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, p.DateNaissance == date)
}

func TestDate(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TABLE t1 (id serial, date date);
	`)
	tu.AssertNoErr(t, err)

	ti := NewDate(2025, 5, 23)
	_, err = db.Exec("INSERT INTO t1 (date) VALUES ($1)", ti)
	tu.AssertNoErr(t, err)

	var v Date
	err = db.QueryRow("SELECT date FROM t1;").Scan(&v)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == ti)

	// JSON
	b, err := json.Marshal(v)
	tu.AssertNoErr(t, err)
	tu.Assert(t, string(b) == `"2025-05-23"`)
	var v2 Date
	err = json.Unmarshal(b, &v2)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == v2)
}

func TestDumpRandomDB(t *testing.T) {
	// t.Skip("generation test: already run")

	noms := [...]string{"kugler", "Méchin", "JOnac", "Martin-Guillard"}
	prenoms := [...]string{"henry", "pierre", "martin", "jean-jacques", "dédé", "Maël"}

	l := make([]Personne, 200)
	for i := range l {
		l[i] = randPersonne()
		if randbool() {
			l[i].Nom = noms[rand.Intn(len(noms))]
		}
		if randbool() {
			l[i].Prenom = prenoms[rand.Intn(len(prenoms))]
		}
	}
	b, err := json.MarshalIndent(l, " ", " ")
	tu.AssertNoErr(t, err)

	err = os.WriteFile("../../controllers/matching/test/samples.json", b, os.ModePerm)
	tu.AssertNoErr(t, err)
}
