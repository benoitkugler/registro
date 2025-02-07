package inscriptions

import (
	"slices"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestController_load(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	const duree = 5
	now := time.Now()
	debut := now.Add((-duree - 2) * 24 * time.Hour) // camp terminé
	_, err := cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(debut), Duree: duree, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(debut), Duree: duree, Ouvert: false}.Insert(db)
	tu.AssertNoErr(t, err)
	c3, err := cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(now), Duree: duree, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(now), Duree: duree, Ouvert: false}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{})

	t.Run("loadCamps", func(t *testing.T) {
		got, err := ct.loadCamps()
		tu.AssertNoErr(t, err)
		tu.Assert(t, slices.Equal(got.IDs(), []cps.IdCamp{c3.Id}))
	})

	t.Run("decodePreinscription", func(t *testing.T) {
		resp, err := pr.Personne{}.Insert(ct.db)
		tu.AssertNoErr(t, err)
		part, err := pr.Personne{}.Insert(ct.db)
		tu.AssertNoErr(t, err)

		pre := preinscription{IdResponsable: resp.Id, IdParticipants: pr.IdPersonneSet{resp.Id: true, part.Id: true}}
		preinsc, err := ct.key.EncryptJSON(pre)
		tu.AssertNoErr(t, err)
		out, err := ct.decodePreinscription(preinsc)
		tu.AssertNoErr(t, err)
		id, err := crypto.DecryptID[pr.IdPersonne](ct.key, out.ResponsablePreIdent)
		tu.AssertNoErr(t, err)
		tu.Assert(t, id == pre.IdResponsable)

		data, err := ct.loadData("", preinsc)
		tu.AssertNoErr(t, err)
		tu.Assert(t, data.InitialInscription.ResponsablePreIdent != "")
	})
}

func TestController_chercheMail(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{})

	got, _ := ct.chercheMail("")
	tu.Assert(t, len(got.responsables) == 0)
	p1, err := pr.Personne{Etatcivil: pr.Etatcivil{Mail: "xx@free.fr", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p2, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p3, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: p1.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = ds.Dossier{IdTaux: 1, IdResponsable: p1.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p2.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p3.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	out, err := ct.chercheMail("xx@free.fr ")
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.responsables) == 1)
	tu.Assert(t, slices.Equal(utils.MapKeysSorted(out.idsParticipants), []pr.IdPersonne{p2.Id, p3.Id}))
}

func loadEnv(t *testing.T) (config.Asso, config.SMTP) {
	tu.LoadEnv(t, "../../env.sh")

	cfg, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	creds, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return cfg, creds
}

func TestController_saveInscription(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	cfg, creds := loadEnv(t)

	camp, err := cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(time.Now()), Duree: 3, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, creds, cfg)

	err = ct.saveInscription("", Inscription{})
	tu.AssertErr(t, err)

	err = ct.saveInscription("localhost", Inscription{
		Responsable: inscriptions.ResponsableLegal{
			Nom: "Kug", Prenom: "Ben",
			DateNaissance: shared.NewDate(2000, 1, 1),
		},
		Participants: []Participant{
			{IdCamp: camp.Id, DateNaissance: shared.Date(time.Now())},
			{IdCamp: camp.Id, DateNaissance: shared.Date(time.Now())},
		},
	})
	tu.AssertNoErr(t, err)
}
