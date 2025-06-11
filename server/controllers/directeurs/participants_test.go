package directeurs

import (
	"os"
	"slices"
	"testing"

	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestFichesSanitaires(t *testing.T) {
	err := pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	err = pr.Fichesanitaire{IdPersonne: pe1.Id, TraitementMedical: true}.Insert(db)
	tu.AssertNoErr(t, err)
	err = pr.Fichesanitaire{IdPersonne: pe2.Id, TraitementMedical: true}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	pa1, err := cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe3.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	fiches, err := ct.loadFichesSanitaires(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(fiches) == 3)

	_, _, err = ct.downloadFicheSanitaire(camp.Id, pa1.Id)
	tu.AssertNoErr(t, err)

	_, _, err = ct.downloadFichesSanitaires(camp.Id)
	tu.AssertNoErr(t, err)
}

func TestMessages(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ev, err := events.Event{IdDossier: dossier.Id, Kind: events.Message}.Insert(db)
	tu.AssertNoErr(t, err)
	err = events.EventMessage{IdEvent: ev.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, asso: asso, smtp: smtp}

	_, err = ct.loadMessages(camp.Id)
	tu.AssertNoErr(t, err)

	out, err := ct.setMessageSeen(camp.Id, ev.Id, true)
	tu.AssertNoErr(t, err)
	tu.Assert(t, slices.Equal(out.Content.VuParCampsIDs, []cps.IdCamp{camp.Id}))
	_, err = ct.setMessageSeen(camp.Id, ev.Id, true)
	tu.AssertNoErr(t, err)
	out, err = ct.setMessageSeen(camp.Id, ev.Id, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, slices.Equal(out.Content.VuParCampsIDs, []cps.IdCamp(nil)))

	_, err = ct.createMessage("", camp.Id, CreateMessageIn{Contenu: "dmlqsd", IdDossier: dossier.Id})
	tu.AssertNoErr(t, err)
}
