package directeurs

import (
	"testing"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func loadEnv(t *testing.T) (config.Asso, config.SMTP) {
	tu.LoadEnv(t, "../../env.sh")

	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return asso, smtp
}

func TestToken(t *testing.T) {
	ct := Controller{key: crypto.Encrypter{}}
	token, err := ct.NewToken(25)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(token) > 10)
}

func TestLoggin(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	directeur, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1, Password: "pass1"}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Equipier{IdCamp: camp.Id, Roles: cps.Roles{cps.Direction}, IdPersonne: directeur.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, password: "globalPass"}

	_, err = ct.loggin("", camp.Id+1, "")
	tu.AssertErr(t, err)

	out, err := ct.loggin("", camp.Id, "wrong")
	tu.AssertNoErr(t, err)
	tu.Assert(t, !out.IsValid)

	out, err = ct.loggin("", camp.Id, "pass1")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.IsValid)

	out, err = ct.loggin("", camp.Id, "globalPass")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.IsValid)

	out, err = ct.loggin("", camp.Id, ct.shortKey.ShortKey(directeur.Id))
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.IsValid)
}
