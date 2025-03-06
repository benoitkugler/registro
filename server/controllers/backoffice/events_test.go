package backoffice

import (
	"testing"
	"time"

	"registro/config"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

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

	ct := Controller{db: db.DB, smtp: smtp, asso: asso}

	d1, err := ct.createDossier(pe1.Id)
	tu.AssertNoErr(t, err)

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
}
