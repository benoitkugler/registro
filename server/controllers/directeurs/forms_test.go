package directeurs

import (
	"testing"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestForms(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	ct := Controller{db: db.DB}

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	pa1, err := cps.Participant{IdPersonne: pe1.Id, IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdPersonne: pe2.Id, IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdPersonne: pe3.Id, IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	form, err := cps.Form{IdCamp: camp.Id, Champs: cps.Champs{
		cps.Champ{Question: cps.ChampQCM{}},
		cps.Champ{Question: cps.ChampTexte{}},
		cps.Champ{Question: cps.ChampTexte{}},
	}}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	err = cps.ParticipantForm{IdParticipant: pa1.Id, IdForm: form.Id, IdCamp: camp.Id, Reponses: cps.FormReponses{
		cps.ChampReponseQCM{},
		cps.ChampReponseTexte(""),
		cps.ChampReponseTexte(""),
	}}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	rep, err := ct.loadFormReponses(form.Id, camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(rep.Inscrits) == 2)
	tu.Assert(t, len(rep.Champs) == 3)
	tu.Assert(t, len(rep.Inscrits[0].Reponses) == 3)
	tu.Assert(t, len(rep.Inscrits[1].Reponses) == 3)
}
