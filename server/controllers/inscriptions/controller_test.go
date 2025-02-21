package inscriptions

import (
	"database/sql"
	"net/url"
	"slices"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	ev "registro/sql/events"
	in "registro/sql/inscriptions"
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

	p1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Equipier{IdCamp: c3.Id, IdPersonne: p1.Id, Roles: cps.Roles{cps.Direction}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{})

	t.Run("loadCamps", func(t *testing.T) {
		camps, tauxs, equipiers, personnes, err := ct.loadCamps()
		tu.AssertNoErr(t, err)
		tu.Assert(t, slices.Equal(camps.IDs(), []cps.IdCamp{c3.Id}))
		tu.Assert(t, slices.Equal(tauxs.IDs(), []ds.IdTaux{1}))
		tu.Assert(t, len(equipiers) == 1 && len(personnes) == 1)
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

	links, err := ct.buildPreinscription("localhost", out)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 1)

	u, err := url.Parse(string(links[0].Lien))
	tu.AssertNoErr(t, err)
	insc, err := ct.decodePreinscription(u.Query().Get(preinscriptionKey))
	tu.AssertNoErr(t, err)
	tu.Assert(t, insc.ResponsablePreIdent != "")
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

	taux2, err := ds.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(time.Now()), Duree: 3, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: taux2.Id, DateDebut: shared.NewDateFrom(time.Now()), Duree: 3, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)
	pers, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pers.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdPersonne: pers.Id, IdCamp: camp.Id, IdTaux: camp.IdTaux, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, creds, cfg)

	err = ct.saveInscription("", Inscription{})
	tu.AssertErr(t, err)

	// insc with inconsitent taux
	err = ct.saveInscription("", Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Kug", Prenom: "Ben",
			DateNaissance: shared.NewDate(2000, 1, 1),
		},
		Participants: []Participant{
			{IdCamp: camp.Id, DateNaissance: shared.Date(time.Now())},
			{IdCamp: camp2.Id, DateNaissance: shared.Date(time.Now())},
		},
	})
	tu.AssertErr(t, err)

	// insc with pre-ident to someone already in the camp
	err = ct.saveInscription("", Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Kug", Prenom: "Ben",
			DateNaissance: shared.NewDate(2000, 1, 1),
		},
		Participants: []Participant{
			{PreIdent: crypto.EncryptID(ct.key, pers.Id), IdCamp: camp.Id, DateNaissance: shared.Date(time.Now())},
		},
	})
	tu.AssertErr(t, err)

	err = ct.saveInscription("localhost", Inscription{
		Responsable: in.ResponsableLegal{
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

func TestController_confirmeInscription(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	cfg, creds := loadEnv(t)

	camp, err := cps.Camp{IdTaux: 1, DateDebut: shared.NewDateFrom(time.Now()), Duree: 3, Ouvert: true}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Groupe{IdCamp: camp.Id, Plage: shared.Plage{
		From:  shared.NewDateFrom(time.Now().Add(-50 * 24 * time.Hour)),
		Duree: 100,
	}}.Insert(db)
	tu.AssertNoErr(t, err)
	pers, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, creds, cfg)

	insc, participants, err := ct.buildInscription(Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Kug", Prenom: "Ben",
			DateNaissance: shared.NewDate(2000, 1, 1),
		},
		ResponsablePreIdent: crypto.EncryptID(ct.key, pers.Id),
		Participants: []Participant{
			{IdCamp: camp.Id, DateNaissance: shared.Date(time.Now()), PreIdent: crypto.EncryptID(ct.key, pers.Id)},
			{IdCamp: camp.Id, DateNaissance: shared.Date(time.Now())},
		},
		Message: "Haha joli !",
	})
	tu.AssertNoErr(t, err)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		insc, err = in.Create(tx, insc, participants)
		return err
	})
	tu.AssertNoErr(t, err)

	dossier, err := ct.confirmeInscription(insc.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, dossier.IsValidated == false)
	tu.Assert(t, dossier.MomentInscription.Equal(insc.DateHeure))

	insc, err = in.SelectInscription(ct.db, insc.Id)
	tu.Assert(t, insc.IsConfirmed == true)

	respo, err := pr.SelectPersonne(ct.db, dossier.IdResponsable)
	tu.AssertNoErr(t, err)
	tu.Assert(t, respo.Publicite.PubEte && respo.Publicite.PubHiver)

	events, err := ev.SelectEventsByIdDossiers(ct.db, dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(events) == 1) //  message

	_, err = ct.confirmeInscription(insc.Id) // already confirmed
	tu.AssertErr(t, err)
}

// This test is a command helper for developping,
// to generate on demande new inscriptions
func TestDebug_createInscription(t *testing.T) {
	t.Skip("dev only test")

	cfg, creds := loadEnv(t)
	sqlCreds, err := config.NewDB()
	tu.AssertNoErr(t, err)
	db, err := sqlCreds.ConnectPostgres()
	tu.AssertNoErr(t, err)

	ct := NewController(db, crypto.Encrypter{}, creds, cfg)

	// assume we already have two camps
	camps, _, _, _, err := ct.loadCamps()
	tu.AssertNoErr(t, err)
	campIds := camps.IDs()
	c1, c2 := campIds[0], campIds[1]

	insc, participants, err := ct.buildInscription(Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Yamina", Prenom: utils.RandString(10, false),
			DateNaissance: shared.NewDate(2000, 1, 1),
			Sexe:          pr.Man,
		},
		Participants: []Participant{
			{IdCamp: c1, DateNaissance: shared.Date(time.Now()), Nom: "Martin", Prenom: "Pierre"},
			{IdCamp: c1, DateNaissance: shared.Date(time.Now()), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
			{IdCamp: c2, DateNaissance: shared.Date(time.Now()), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
		},
		Message: utils.RandString(30, true) + "\n" + utils.RandString(10, true),
	})
	tu.AssertNoErr(t, err)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		insc, err = in.Create(tx, insc, participants)
		return err
	})
	tu.AssertNoErr(t, err)
	_, err = ct.confirmeInscription(insc.Id)
	tu.AssertNoErr(t, err)
}
