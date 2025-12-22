package logic

import (
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSondages(t *testing.T) {
	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.NewDate(2000, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.NewDate(2001, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.NewDate(2002, time.January, 5)}}.Insert(db)
	tu.AssertNoErr(t, err)

	d1, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	d2, err := ds.Dossier{IdTaux: 1, IdResponsable: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: d1.Id, IdPersonne: pe1.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: d1.Id, IdPersonne: pe2.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdTaux: 1, IdDossier: d2.Id, IdPersonne: pe3.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Sondage{IdCamp: camp.Id, IdDossier: d1.Id, ReponseSondage: cps.ReponseSondage{InfosAvantSejour: 3, Hebergement: 1}}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Sondage{IdCamp: camp.Id, IdDossier: d2.Id, ReponseSondage: cps.ReponseSondage{InfosAvantSejour: 3, Hebergement: 2}}.Insert(db)
	tu.AssertNoErr(t, err)

	l, err := LoadSondages(db, []cps.IdCamp{camp.Id})
	tu.AssertNoErr(t, err)
	out := l.For(camp.Id)
	tu.Assert(t, len(out.Sondages) == 2)
	tu.Assert(t, out.Moyennes.InfosAvantSejour == 3)
	tu.Assert(t, out.Moyennes.Hebergement == 1.5)
}
