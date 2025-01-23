package personnes

import (
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
	tu.Assert(t, time.Time(p.DateNaissance).Equal(time.Time(date)))
}
