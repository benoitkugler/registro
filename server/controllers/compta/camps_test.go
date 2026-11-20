package compta

import (
	"registro/sql/compta"
	tu "registro/utils/testutils"
	"testing"
)

func TestGetCamps(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	ct := Controller{db: db.DB}
	l, err := ct.getCamps("localhost")
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 2)
	tu.Assert(t, l[0].Compta.Id == 0 && l[1].Compta.Id == 0)

	url := DirecteurURL(ct.key, "localhost", l[0].Id)
	tu.Assert(t, url != "")

	cc, err := ct.loadOrCreateCompta(c1.Id)
	tu.AssertNoErr(t, err)

	err = ct.updateCompta(ComptaExt{Id: cc.Id, Account: compta.Bergerie})
	tu.AssertNoErr(t, err)

	err = ct.deleteCompta(cc.Id)
	tu.AssertNoErr(t, err)
}
