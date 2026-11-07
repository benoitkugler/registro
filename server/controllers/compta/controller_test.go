package compta

import (
	"testing"

	"registro/crypto"
	tu "registro/utils/testutils"
)

func TestConnectionLink(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	c1, _ := createCamps(t, db)

	enc := crypto.NewEncrypter("dummy")
	ct := Controller{db: db.DB, key: enc, adminPassword: "HAHA"}

	lAdmin, err := ct.loggin(true, "HAHA")
	tu.AssertNoErr(t, err)
	tu.Assert(t, lAdmin.IdCompta == -1)

	key := crypto.EncryptID(enc, c1.Id)
	lDir1, err := ct.loggin(false, key)
	tu.AssertNoErr(t, err)
	tu.Assert(t, lDir1.IdCompta > 0)

	lDir2, err := ct.loggin(false, key)
	tu.AssertNoErr(t, err)
	tu.Assert(t, lDir2.IsValid)
	tu.Assert(t, lDir2.IdCompta == lDir1.IdCompta)

	// check login again with token works
	lA2, err := ct.loggin(true, lAdmin.Token)
	tu.AssertNoErr(t, err)
	tu.Assert(t, lA2.IsValid && lA2.IdCompta == userAdmin)

	l3, err := ct.loggin(false, lDir1.Token)
	tu.AssertNoErr(t, err)
	tu.Assert(t, l3.IsValid && l3.IdCompta == lDir1.IdCompta)
	_, err = ct.loggin(true, lDir1.Token) // no admin rights !
	tu.AssertErr(t, err)

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
