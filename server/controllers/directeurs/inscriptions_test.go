package directeurs

import (
	"testing"
	"time"

	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func Test_inscriptions(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	forAge7 := shared.NewDateFrom(tu.DateFor(7))

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{Nom: "test", DateNaissance: forAge7}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false, Identite: pr.Identite{Nom: "test", DateNaissance: forAge7}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12, Nom: "C2", DateDebut: shared.Date(time.Now())}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 8, AgeMax: 12, Nom: "C3", DateDebut: shared.Date(time.Now())}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	pa3, err := cps.Participant{IdCamp: camp2.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1, Statut: cps.AStatuer}.Insert(db)
	tu.AssertNoErr(t, err)

	asso, smtp := loadEnv(t)
	ct := Controller{db: db.DB, asso: asso, smtp: smtp}

	t.Run("load", func(t *testing.T) {
		out, err := ct.getInscriptions(camp1.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(out.Inscriptions) == 0 && out.PendingCount == 0) // all participant for camp1 are validated

		out, err = ct.getInscriptions(camp2.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(out.Inscriptions) == 1)
		insc := out.Inscriptions[0]
		tu.Assert(t, insc.Participants[0].Camp.Id == camp2.Id)
		tu.Assert(t, insc.Participants[1].Camp.Id == camp1.Id)
		tu.Assert(t, insc.Participants[2].Camp.Id == camp1.Id)
	})

	t.Run("valide", func(t *testing.T) {
		hints, err := logic.HintValideInscription(ct.db, directeursBypass, dossier1.Id)
		tu.AssertNoErr(t, err)

		values := make(map[cps.IdParticipant]cps.StatutParticipant)
		for k, v := range hints {
			values[k] = v.Statut
		}
		err = ct.valideInscription("localhost", InscriptionsValideIn{
			IdDossier: dossier1.Id,
			Statuts:   values,
			SendMail:  true,
		}, camp1.Id)
		tu.AssertNoErr(t, err)
		data, err := logic.LoadDossier(db, dossier1.Id)
		tu.AssertNoErr(t, err)

		out, err := ct.getInscriptions(camp1.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(out.Inscriptions) == 0)

		err = ct.valideInscription("localhost", InscriptionsValideIn{
			IdDossier: dossier1.Id,
			Statuts:   values,
		}, camp2.Id)
		tu.AssertNoErr(t, err)
		data, err = logic.LoadDossier(db, dossier1.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, data.Participants[pa3.Id].Statut == cps.AStatuer) // not validable because of age

		out, err = ct.getInscriptions(camp2.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(out.Inscriptions) == 1)
	})
}
