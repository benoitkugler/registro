package backoffice

import (
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestController_participants(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	taux2, err := ds.Taux{Label: "autre", Euros: 1000, FrancsSuisse: 1560}.Insert(db)
	tu.AssertNoErr(t, err)

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: taux2.Id, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)
	camp3, err := cps.Camp{IdTaux: taux2.Id, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	structure, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	asso, smtp := loadEnv(t)
	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir()), smtp: smtp, asso: asso}

	part, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp1.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)

	part.Participant.Statut = cps.Inscrit
	part.Participant.QuotientFamilial = 48
	err = ct.updateParticipant(part.Participant)
	tu.AssertNoErr(t, err)

	aide, err := ct.createAide(AidesCreateIn{IdParticipant: part.Participant.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)
	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test.png")
	tu.AssertNoErr(t, err)

	files, err := fs.SelectAllFiles(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 1)

	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp2.Id, IdPersonne: pe1.Id})
	tu.AssertErr(t, err) // inconsistent taux

	err = ct.deleteParticipant(part.Participant.Id)
	tu.AssertNoErr(t, err)

	files, err = fs.SelectAllFiles(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 0)

	p2, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp2.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err) // now the change of taux is OK

	err = ct.moveParticipant(ParticipantsMoveIn{Id: p2.Participant.Id, Target: camp1.Id})
	tu.AssertErr(t, err) // invalid taux

	err = ct.moveParticipant(ParticipantsMoveIn{Id: p2.Participant.Id, Target: camp2.Id})
	tu.AssertErr(t, err) // same camp

	p3, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp3.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)

	err = ct.moveParticipant(ParticipantsMoveIn{Id: p2.Participant.Id, Target: camp3.Id})
	tu.AssertErr(t, err) // already in camp

	err = ct.deleteParticipant(p3.Participant.Id)
	tu.AssertNoErr(t, err)

	err = ct.moveParticipant(ParticipantsMoveIn{Id: p2.Participant.Id, Target: camp3.Id})
	tu.AssertNoErr(t, err)

	_, err = ct.setPlaceLiberee("localhost", p2.Participant.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.setPlaceLiberee("localhost", p2.Participant.Id)
	tu.AssertErr(t, err) // already notified !
}
