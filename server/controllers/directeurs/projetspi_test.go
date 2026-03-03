package directeurs

import (
	"testing"

	cps "registro/sql/camps"
	tu "registro/utils/testutils"
)

func TestProjetSpi(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	err = ct.updateProjetSpi(cps.ProjetSpi{IdCamp: camp1.Id, VisiteLibrairie: cps.Oui, Description: "ùmlzm"})
	tu.AssertNoErr(t, err)
}
