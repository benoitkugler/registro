package dons

import (
	"testing"
	"time"

	"registro/config"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestLoggin(t *testing.T) {
	ct := Controller{password: "1234"}
	out, err := ct.loggin("smldsmld")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out == LogginOut{})

	out, err = ct.loggin("1234")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.IsValid && len(out.Token) > 0)
}

func TestIdentifieDon(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "Kugler", Prenom: "Benoit", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "Kugler", Prenom: "Estelle", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	id, err := ct.identifieAddDon(dons.Don{}, pr.Etatcivil{})
	tu.AssertNoErr(t, err)
	tu.Assert(t, id != pe1.Id && id != pe2.Id)

	id, err = ct.identifieAddDon(dons.Don{}, pr.Etatcivil{Nom: "Kugler", Prenom: "Benoit", DateNaissance: shared.NewDate(2000, 1, 1)})
	tu.AssertNoErr(t, err)
	tu.Assert(t, id == pe1.Id)
}

func loadEnv(t *testing.T) (config.Asso, config.SMTP) {
	tu.LoadEnv(t, "../../env.sh")

	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return asso, smtp
}

func TestDonsAPI(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "Kugler", Prenom: "Benoit", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := dons.Organisme{Nom: "Eglise chrétienne"}.Insert(db)
	tu.AssertNoErr(t, err)

	asso, smtp := loadEnv(t)
	ct := Controller{db: db.DB, asso: asso, smtp: smtp}

	_, err = ct.createDon(dons.Don{IdPersonne: pe1.Id.Opt(), Montant: ds.NewEuros(50.1), ModePaiement: ds.Cheque, Date: shared.NewDateFrom(time.Now())})
	tu.AssertNoErr(t, err)
	_, err = ct.createDon(dons.Don{IdOrganisme: pe2.Id.Opt(), Montant: ds.NewFrancsuisses(50.1), ModePaiement: ds.Cheque, Date: shared.NewDateFrom(time.Now())})
	tu.AssertNoErr(t, err)

	out, err := ct.loadDons()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dons) == 2)
	tu.Assert(t, len(out.YearTotals) == 1)

	excel, err := ct.exportDonsExcel(time.Now().Year())
	tu.Write(t, "Dons.xlsx", excel)
}
