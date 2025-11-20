package services

import (
	"testing"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSearchMail(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, smtp, asso)

	got, err := ct.searchMailAndSend("", "")
	tu.AssertNoErr(t, err)
	tu.Assert(t, got.Found == 0)

	respo, err := pr.Personne{Etatcivil: pr.Etatcivil{Mail: "xx@free.fr", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p2, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p3, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: respo.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = ds.Dossier{IdTaux: 1, IdResponsable: respo.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p2.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p3.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	out, err := ct.searchMailAndSend("", "xx@free.fr")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.Found == 2)
}
