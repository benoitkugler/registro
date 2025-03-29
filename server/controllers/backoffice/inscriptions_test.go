package backoffice

import (
	"testing"
	"time"

	"registro/controllers/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestController_getInscriptions(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp2.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp2.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now().Add(time.Hour)}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now().Add(time.Hour), IsValidated: true}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	out, err := ct.getInscriptions()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out) == 2)
	tu.Assert(t, out[0].Dossier.Id == dossier1.Id)
}

func TestValideInscription(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	hints, err := ct.hintValideInscription(dossier1.Id)
	tu.AssertNoErr(t, err)

	values := make(map[cps.IdParticipant]cps.ListeAttente)
	for k, v := range hints {
		values[k] = v.Statut
	}
	err = ct.valideInscription(InscriptionsValideIn{IdDossier: dossier1.Id, Statuts: values})
	tu.AssertNoErr(t, err)

	data, err := logic.LoadDossier(db, dossier1.Id)
	tu.AssertNoErr(t, err)
	for _, part := range data.Participants {
		tu.Assert(t, part.Statut == cps.AttenteProfilInvalide)
	}
	tu.Assert(t, data.Dossier.IsValidated)

	err = ct.deleteDossier(data.Dossier.Id)
	tu.AssertNoErr(t, err)
}
