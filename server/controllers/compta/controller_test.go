package compta

import (
	"registro/crypto"
	tu "registro/utils/testutils"
	"testing"
)

func TestConnectionLink(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	enc := crypto.NewEncrypter("dummy")
	ct := Controller{db: db.DB, key: enc, adminPassword: "HAHA"}

	logged, err := ct.loggin(true, "HAHA")
	tu.AssertNoErr(t, err)
	tu.Assert(t, logged.IdCompta == -1)

	key := crypto.EncryptID(enc, c1.Id)
	logged, err = ct.loggin(false, key)
	tu.AssertNoErr(t, err)
	tu.Assert(t, logged.IdCompta > 0)

	logged2, err := ct.loggin(false, key)
	tu.AssertNoErr(t, err)
	tu.Assert(t, logged.IdCompta > 0)
	tu.Assert(t, logged2.IdCompta == logged.IdCompta)

	out, err := ct.loggin(true, "XXX")
	tu.AssertNoErr(t, err)
	tu.Assert(t, !out.IsValid)
	_, err = ct.loggin(false, "XXX")
	tu.AssertErr(t, err)
}

func TestAdminToken(t *testing.T) {
	ct := Controller{key: crypto.NewEncrypter("dummy")}
	_, err := ct.newToken(userAdmin)
	tu.AssertNoErr(t, err)
}
