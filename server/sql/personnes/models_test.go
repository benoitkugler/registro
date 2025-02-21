package personnes

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "gen_create.sql")
	defer db.Remove()

	p := randPersonne()
	p, err := p.Insert(db)
	tu.AssertNoErr(t, err)

	date := shared.NewDate(2025, time.September, 23)
	p.DateNaissance = date
	_, err = p.Update(db)
	tu.AssertNoErr(t, err)

	p, err = SelectPersonne(db, p.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, p.DateNaissance == date)

	// check that FS is not allowed on temp people
	p.IsTemp = true
	_, err = p.Update(db)
	tu.AssertNoErr(t, err)

	err = Fichesanitaire{IdPersonne: p.Id}.Insert(db)
	tu.AssertErr(t, err)

	p.IsTemp = false
	_, err = p.Update(db)
	tu.AssertNoErr(t, err)

	err = Fichesanitaire{IdPersonne: p.Id}.Insert(db)
	tu.AssertNoErr(t, err)
}

func TestDumpRandomDB(t *testing.T) {
	t.Skip("generation test: already run")

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
	l[0].IsTemp = false

	b, err := json.MarshalIndent(l, " ", " ")
	tu.AssertNoErr(t, err)

	err = os.WriteFile("../../controllers/search/test/samples.json", b, os.ModePerm)
	tu.AssertNoErr(t, err)
}
