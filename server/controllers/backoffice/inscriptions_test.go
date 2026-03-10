package backoffice

import (
	"testing"
	"time"

	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	in "registro/sql/inscriptions"
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

	ct := Controller{db: db.DB}

	out, err := ct.getInscriptions(false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out) == 1)
	tu.Assert(t, out[0].Dossier.Id == dossier1.Id)
}

func TestValideInscription(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{Nom: "melzmel", Prenom: "szlùs", DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false, Identite: pr.Identite{Nom: "melzmel", Prenom: "szlsdsmldlmsùs", DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12, Nom: "Séjour Test"}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12, Nom: "Séjour Test"}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp2.Id, IdPersonne: pe2.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, asso: asso, smtp: smtp}

	hints, err := logic.HintValideInscription(ct.db, backofficeRights, dossier1.Id)
	tu.AssertNoErr(t, err)

	values := make(map[cps.IdParticipant]cps.StatutParticipant)
	for k, v := range hints {
		values[k] = v.Statut
	}
	err = ct.valideInscription("localhost:1323", logic.InscriptionsValideIn{IdDossier: dossier1.Id, Statuts: values, SendMail: true})
	tu.AssertNoErr(t, err)

	data, err := logic.LoadDossier(db, dossier1.Id)
	tu.AssertNoErr(t, err)
	for _, part := range data.Participants {
		tu.Assert(t, part.Statut == cps.AttenteProfilInvalide)
	}

	tu.Assert(t, len(logic.EventsBy[logic.ValidationEvt](data.Events)) == 2) // 2 camps

	err = ct.deleteDossier(data.Dossier.Id)
	tu.AssertNoErr(t, err)
}

func TestSearchDoublons(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp1, err := cps.Camp{IdTaux: 1, DateDebut: shared.Date(time.Now()), Duree: 2, Statut: cps.Ouvert}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1, DateDebut: shared.Date(time.Now()), Duree: 2, Statut: cps.Ouvert}.Insert(db)
	tu.AssertNoErr(t, err)

	i1, err := in.Inscription{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	i2, err := in.Inscription{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	err = in.InscriptionParticipant{Nom: "ugler", Prenom: "benoît", DateNaissance: shared.NewDate(2000, 2, 2), IdInscription: i1.Id, IdCamp: camp1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = in.InscriptionParticipant{Nom: "Ugler", Prenom: "benoit", DateNaissance: shared.NewDate(2000, 2, 2), IdInscription: i1.Id, IdCamp: camp1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = in.InscriptionParticipant{Nom: "ugl er", Prenom: "benoît ", DateNaissance: shared.NewDate(2000, 2, 2), IdInscription: i1.Id, IdCamp: camp2.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = in.InscriptionParticipant{Nom: "ugler", Prenom: "benoît", DateNaissance: shared.NewDate(2000, 2, 2), IdInscription: i2.Id, IdCamp: camp1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = in.InscriptionParticipant{Nom: "ugler", Prenom: "benoît", DateNaissance: shared.NewDate(2000, 2, 2), IdInscription: i2.Id, IdCamp: camp1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = in.InscriptionParticipant{Nom: "ugler", Prenom: "benoît", DateNaissance: shared.NewDate(2001, 2, 2), IdInscription: i2.Id, IdCamp: camp2.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}
	out, err := ct.searchInscriptionsDoublons()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Participants) == 1 && len(out.Participants[0]) == 5)
}
