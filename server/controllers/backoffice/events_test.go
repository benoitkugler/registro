package backoffice

import (
	"testing"
	"time"

	"registro/config"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

func createMessage(db events.DB, idDossier ds.IdDossier, origine events.MessageOrigine, origineCamp events.OptIdCamp) error {
	return events.CreateMessage(db, idDossier, time.Now(), utils.RandString(30, true), origine, origineCamp)
}

func Test_messages(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)

	pe1, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now()), Prenom: "Benoit"}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, smtp: smtp, asso: asso}

	d1, err := ct.createDossier(pe1.Id)
	tu.AssertNoErr(t, err)

	t.Run("message CRUD", func(t *testing.T) {
		event, err := ct.sendMessage("localhost", EventsSendMessageIn{IdDossier: d1.Id, Contenu: `
		Merci pur l'inscriptio !
		
		Question : donnez moi vos on cafs
		
		Merci
		Marie-Pierre
		`})
		tu.AssertNoErr(t, err)
		tu.Assert(t, event.Kind == events.Message)

		err = ct.deleteEvent(event.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("mark seen", func(t *testing.T) {
		err = createMessage(ct.db, d1.Id, events.FromEspaceperso, events.OptIdCamp{})
		tu.AssertNoErr(t, err)
		err = createMessage(ct.db, d1.Id, events.FromBackoffice, events.OptIdCamp{})
		tu.AssertNoErr(t, err)
		err = createMessage(ct.db, d1.Id, events.FromDirecteur, camp1.Id.Opt())
		tu.AssertNoErr(t, err)

		err = ct.markMessagesSeen(d1.Id)
		tu.AssertNoErr(t, err)
	})
}
