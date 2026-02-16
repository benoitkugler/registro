package directeurs

import (
	"os"
	"slices"
	"testing"
	"time"

	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/sql/shared"
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

	err = pr.Fichesanitaire{IdPersonne: pe1.Id, TraitementMedical: "Il doit prendre des méicatments !"}.Insert(db)
	tu.AssertNoErr(t, err)
	err = pr.Fichesanitaire{IdPersonne: pe2.Id, AllergiesAlimentaires: "Le mais !"}.Insert(db)
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

	_, _, err = ct.renderFichesSanitaires(camp.Id)
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

func TestGroupes(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.NewDate(2000, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.NewDate(2001, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.NewDate(2002, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	pa1, err := cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe1.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	pa2, err := cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe2.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	pa3, err := cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: dossier.Id, IdPersonne: pe3.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	out, err := ct.getGroupes(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Groupes) == 0 && len(out.ParticipantsToGroupe) == 0)

	groupe1, err := ct.createGroupe(camp.Id)
	tu.AssertNoErr(t, err)
	groupe2, err := ct.createGroupe(camp.Id)
	tu.AssertNoErr(t, err)
	groupe3, err := ct.createGroupe(camp.Id)
	tu.AssertNoErr(t, err)

	groupe1.Couleur = "#457898"
	err = ct.updateGroupe(camp.Id, groupe1)
	tu.AssertNoErr(t, err)

	err = ct.setParticipantGroupe(camp.Id, pa1.Id, groupe1.Id)
	tu.AssertNoErr(t, err)
	err = ct.setParticipantGroupe(camp.Id, pa2.Id, groupe2.Id)
	tu.AssertNoErr(t, err)

	err = ct.updateGroupesPlages(camp.Id, UpdateFinsIn{map[cps.IdGroupe]shared.Date{
		groupe1.Id: shared.NewDate(1990, time.January, 5),
		groupe2.Id: shared.NewDate(1990, time.January, 5),
		groupe3.Id: shared.NewDate(2003, time.January, 5),
	}, false})
	tu.AssertNoErr(t, err)

	out, err = ct.getGroupes(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Groupes) == 3 && len(out.ParticipantsToGroupe) == 3)
	tu.Assert(t, out.ParticipantsToGroupe[pa1.Id].IdGroupe == groupe1.Id)
	tu.Assert(t, out.ParticipantsToGroupe[pa2.Id].IdGroupe == groupe2.Id)
	tu.Assert(t, out.ParticipantsToGroupe[pa3.Id].IdGroupe == groupe3.Id)

	err = ct.updateGroupesPlages(camp.Id, UpdateFinsIn{map[cps.IdGroupe]shared.Date{
		groupe1.Id: shared.NewDate(1990, time.January, 5),
		groupe2.Id: shared.NewDate(1990, time.January, 5),
		groupe3.Id: shared.NewDate(2003, time.January, 5),
	}, true})
	tu.AssertNoErr(t, err)

	out, err = ct.getGroupes(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.ParticipantsToGroupe[pa1.Id].IdGroupe == groupe3.Id)
	tu.Assert(t, out.ParticipantsToGroupe[pa2.Id].IdGroupe == groupe3.Id)

	err = ct.setParticipantGroupe(camp.Id, pa2.Id, 0)
	tu.AssertNoErr(t, err)

	out, err = ct.getGroupes(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.ParticipantsToGroupe[pa2.Id].IdGroupe == 0)

	err = ct.deleteGroupe(camp.Id, groupe1.Id)
	tu.AssertNoErr(t, err)
}
