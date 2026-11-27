package backoffice

import (
	"fmt"
	"testing"
	"time"

	"registro/config"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	fs "registro/sql/files"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestOffuscateur(t *testing.T) {
	offuscateur := newOffuscateur[int64]("IN", 8, 3)
	for id := int64(0); id < 500_000; id++ {
		res, ok := offuscateur.Unmask(offuscateur.Mask(id))
		tu.Assert(t, ok)
		tu.Assert(t, res == id)
	}
	fmt.Println(offuscateur.Mask(1))
	fmt.Println(offuscateur.Mask(456))
	fmt.Println(offuscateur.Mask(15456))
}

func TestController_searchDossiers(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12, Prix: ds.NewEuros(100)}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier2, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp1.Id, IdDossier: dossier2.Id, IdTaux: 1, IdPersonne: pe3.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdDossier: dossier1.Id, IdTaux: 1, IdPersonne: pe1.Id, Statut: cps.AttenteCampComplet}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdDossier: dossier1.Id, IdTaux: 1, IdPersonne: pe2.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ds.Paiement{Montant: ds.NewEuros(50), IdDossier: dossier1.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = createMessage(db, dossier1.Id, events.Espaceperso, cps.OptIdCamp{})
	tu.AssertNoErr(t, err)
	err = createMessage(db, dossier1.Id, events.Directeur, camp1.Id.Opt())
	tu.AssertNoErr(t, err)
	err = createMessage(db, dossier2.Id, events.Espaceperso, cps.OptIdCamp{})
	tu.AssertNoErr(t, err)
	err = createMessage(db, dossier2.Id, events.Backoffice, cps.OptIdCamp{})
	tu.AssertNoErr(t, err)
	err = createMessage(db, dossier2.Id, events.Backoffice, cps.OptIdCamp{})
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	out, err := ct.searchDossiers(SearchDossierIn{Pattern: OffuscateurVirements.Mask(dossier1.Id)}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 1)

	out, err = ct.searchDossiers(SearchDossierIn{Pattern: fmt.Sprintf("id:%d", dossier2.Id)}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 1)

	out, err = ct.searchDossiers(SearchDossierIn{Pattern: fmt.Sprintf("id:%d", dossier2.Id+1)}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 0)

	out, err = ct.searchDossiers(SearchDossierIn{Pattern: "test"}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 0)

	out, err = ct.searchDossiers(SearchDossierIn{SortByNewMessages: true}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 2)
	tu.Assert(t, out.Dossiers[0].Id == dossier1.Id) // 2 messages

	out, err = ct.searchDossiers(SearchDossierIn{Reglement: Partiel, Attente: AvecAttente}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 1)
	out, err = ct.searchDossiers(SearchDossierIn{Reglement: Total}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 0)
	out, err = ct.searchDossiers(SearchDossierIn{Reglement: Total, Attente: AvecAttente}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 0)
}

func TestController_searchDossiersNotifications(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier2, err := ds.Dossier{IdResponsable: pe2.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier3, err := ds.Dossier{IdResponsable: pe3.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	part1, err := cps.Participant{IdCamp: camp.Id, IdDossier: dossier1.Id, IdTaux: 1, IdPersonne: pe1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier2.Id, IdTaux: 1, IdPersonne: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier3.Id, IdTaux: 1, IdPersonne: pe3.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	// justificatif d'aide en attente
	file1, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileAide{IdFile: file1.Id, IdParticipant: part1.Id}).Insert(db)
	tu.AssertNoErr(t, err)

	// attestation QF en attente
	file2, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileQuotientFamilial{IdFile: file2.Id, IdDossier: dossier2.Id}).Insert(db)
	tu.AssertNoErr(t, err)

	// attestation QF déjà validée : ce n'est plus une notification
	file3, err := fs.File{ReviewStatus: pr.FileReviewValidated}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileQuotientFamilial{IdFile: file3.Id, IdDossier: dossier3.Id}).Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	out, err := ct.searchDossiers(SearchDossierIn{}, false)
	tu.AssertNoErr(t, err)
	notifications := make(map[ds.IdDossier]int, len(out.Dossiers))
	for _, header := range out.Dossiers {
		notifications[header.Id] = header.Notifications
	}
	tu.Assert(t, notifications[dossier1.Id] == 1)
	tu.Assert(t, notifications[dossier2.Id] == 1)
	tu.Assert(t, notifications[dossier3.Id] == 0)

	out, err = ct.searchDossiers(SearchDossierIn{OnlyNotifications: true}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 2)
	for _, header := range out.Dossiers {
		tu.Assert(t, header.Id != dossier3.Id)
	}
}

func TestController_aides(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	part, err := cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	structure, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir())}

	aide, err := ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)

	// au plus une aide par participant
	_, err = ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertErr(t, err)

	// check structure is not removable now
	err = ct.deleteStructureaide(structure.Id)
	tu.AssertErr(t, err)

	aide.Valeur = ds.NewEuros(26)
	aide.ParJour = true
	err = ct.updateAide(aide)
	tu.AssertNoErr(t, err)

	file, err := ct.uploadAideJustificatif(aide.Id, tu.PngData, "test1.png")
	tu.AssertNoErr(t, err)
	tu.Assert(t, file.Key != "")

	_, err = ct.uploadAideJustificatif(aide.Id, tu.PngData, "test2.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteAideJustificatif(aide.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.uploadAideJustificatif(aide.Id, tu.PngData, "test3.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteAide(aide.Id)
	tu.AssertNoErr(t, err)

	aide, err = ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)

	_, err = ct.uploadAideJustificatif(aide.Id, tu.PngData, "test3.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteDossier(dossier1.Id)
	tu.AssertNoErr(t, err)

	// check structure is removable now
	err = ct.deleteStructureaide(structure.Id)
	tu.AssertNoErr(t, err)
}

func TestController_aideJustificatifAttente(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdResponsable: pe.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	part, err := cps.Participant{IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	structure, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir())}

	// justificatif téléversé par les parents : rattaché au participant, sans aide
	file, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileAide{IdFile: file.Id, IdParticipant: part.Id}).Insert(db)
	tu.AssertNoErr(t, err)
	_, err = fs.UploadFile(ct.files, db, file, tu.PngData, "attente.png")
	tu.AssertNoErr(t, err)

	aide, err := ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)

	link, found, err := fs.SelectFileAideByIdFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found)
	tu.Assert(t, link.IdParticipant == part.Id)
	tu.Assert(t, link.IdAide.Is(aide.Id))
	file, err = fs.SelectFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, file.ReviewStatus == pr.FileReviewValidated)

	// remplacer le justificatif ne réinitialise pas la validation
	_, err = ct.uploadAideJustificatif(aide.Id, tu.PngData, "remplace.png")
	tu.AssertNoErr(t, err)
	file, err = fs.SelectFile(db, file.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, file.ReviewStatus == pr.FileReviewValidated)

	// un justificatif en attente est supprimé avec le participant
	pe2, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	part2, err := cps.Participant{IdCamp: camp.Id, IdPersonne: pe2.Id, IdDossier: dossier.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	file2, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = (fs.FileAide{IdFile: file2.Id, IdParticipant: part2.Id}).Insert(db)
	tu.AssertNoErr(t, err)
	_, err = fs.UploadFile(ct.files, db, file2, tu.PngData, "attente2.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteParticipant(part2.Id)
	tu.AssertNoErr(t, err)
	links, err := fs.SelectFileAidesByIdParticipants(db, part2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 0)
}

func TestController_paiements(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir())}

	out, err := ct.createPaiement(false, dossier1.Id)
	tu.AssertNoErr(t, err)

	out.Montant.Currency = ds.FrancsSuisse
	err = ct.updatePaiement(out)
	tu.AssertErr(t, err) // invalid currency

	out.Montant = ds.NewEuros(56.5)
	err = ct.updatePaiement(out)
	tu.AssertNoErr(t, err) // invalid currency

	err = ct.deleteDossier(dossier1.Id)
	tu.AssertNoErr(t, err)
}

func loadEnv(t *testing.T) (config.Asso, config.SMTP) {
	tu.LoadEnv(t, "../../env.sh")

	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return asso, smtp
}

func TestController_mergeDossiers(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	pe1, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir()), smtp: smtp, asso: asso}

	d1, err := ct.createDossier(pe1.Id)
	tu.AssertNoErr(t, err)
	d2, err := ct.createDossier(pe2.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: d1.Id, IdPersonne: pe1.Id, IdCamp: camp1.Id})
	tu.AssertNoErr(t, err)
	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: d2.Id, IdPersonne: pe2.Id, IdCamp: camp1.Id})
	tu.AssertNoErr(t, err)

	in1, err := in.Inscription{IdTaux: 1, ConfirmedAsDossier: d2.Id.Opt()}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = ct.createPaiement(false, d2.Id)
	tu.AssertNoErr(t, err)

	_, err = events.Event{Kind: events.Message, IdDossier: d2.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = events.Event{Kind: events.Statuation, IdDossier: d2.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = events.Event{Kind: events.PlaceLiberee, IdDossier: d2.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = ct.mergeDossier("", DossiersMergeIn{d2.Id, d1.Id, true})
	tu.AssertNoErr(t, err)

	merged, err := ct.loadDossier("", d1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(merged.Dossier.Events) == 2)

	in1, err = in.SelectInscription(db, in1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, in1.ConfirmedAsDossier == d1.Id.Opt())
}

func TestQueryReglement(t *testing.T) {
	tu.Assert(t, Partiel.match(logic.EnCours))
	tu.Assert(t, !Partiel.match(logic.NonCommence))
	tu.Assert(t, (Partiel | Zero).match(logic.EnCours))
	tu.Assert(t, (Partiel | Zero).match(logic.NonCommence))
	tu.Assert(t, !(Partiel | Zero).match(logic.Complet))
}

func TestEstimeRemises(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Identite: pr.Identite{Nom: "Kugler"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{Nom: "kugler"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2bis, err := pr.Personne{Identite: pr.Identite{Nom: "kugler"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{Identite: pr.Identite{Nom: "hug"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3bis, err := pr.Personne{Identite: pr.Identite{Nom: "hug"}}.Insert(db)
	tu.AssertNoErr(t, err)
	eq, err := pr.Personne{Identite: pr.Identite{Nom: "hug", Ville: "v3", DateNaissance: shared.Date(tu.DateFor(20))}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe5, err := pr.Personne{Identite: pr.Identite{Nom: "sansfamille"}}.Insert(db)
	tu.AssertNoErr(t, err)

	respo1, err := pr.Personne{Identite: pr.Identite{Ville: "Begude"}}.Insert(db)
	tu.AssertNoErr(t, err)
	respo2, err := pr.Personne{Identite: pr.Identite{Ville: "Lyon"}}.Insert(db)
	tu.AssertNoErr(t, err)
	respo3, err := pr.Personne{Identite: pr.Identite{Ville: "v3"}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, asso: config.Asso{ID: "acve"}}
	ct.asso.RemisesHints = config.RemisesHints{ParentEquipier: 5, AutresInscrits: 6}

	camp1, err := ct.createCamp("localhost")
	tu.AssertNoErr(t, err)
	idTaux := camp1.Taux.Id
	idCamp := camp1.Camp.Camp.Id

	d1, err := ct.createDossier(respo1.Id)
	tu.AssertNoErr(t, err)
	d2, err := ct.createDossier(respo2.Id)
	tu.AssertNoErr(t, err)
	d3, err := ct.createDossier(respo3.Id)
	tu.AssertNoErr(t, err)

	// famille 1
	_, err = cps.Participant{IdDossier: d1.Id, IdPersonne: pe1.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdDossier: d1.Id, IdPersonne: pe2.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	// famille 2
	_, err = cps.Participant{IdDossier: d2.Id, IdPersonne: pe2bis.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	// famille 3
	_, err = cps.Participant{IdDossier: d3.Id, IdPersonne: pe3.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdDossier: d3.Id, IdPersonne: pe3bis.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ct.createEquipier(CreateEquipierIn{IdPersonne: eq.Id, IdCamp: idCamp})
	tu.AssertNoErr(t, err)

	// autre
	_, err = cps.Participant{IdDossier: d2.Id, IdPersonne: pe5.Id, IdCamp: idCamp, IdTaux: idTaux, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	out, err := ct.estimeRemises(ct.db, []cps.IdCamp{idCamp})
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(out) == 4)
	tu.Assert(t, out[0].IdParticipant == 1 && out[0].Hint == cps.Remises{Famille: 6})
	tu.Assert(t, out[1].IdParticipant == 2 && out[1].Hint == cps.Remises{Famille: 6})
	tu.Assert(t, out[2].IdParticipant == 4 && out[2].Hint == cps.Remises{Equipiers: 5, Famille: 6})
	tu.Assert(t, out[3].IdParticipant == 5 && out[3].Hint == cps.Remises{Equipiers: 5, Famille: 6})

	err = ct.applyRemises(out[1:2])
	tu.AssertNoErr(t, err)
}

func TestController_soldeDossier(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Identite: pr.Identite{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, files: fs.NewFileSystem(t.TempDir())}
	camp1, err := ct.createCamp("localhost")
	tu.AssertNoErr(t, err)
	part, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp1.Camp.Camp.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)

	// start with non empty remise
	part.Participant.Statut = cps.Inscrit
	part.Participant.Remises.Speciale = cps.Montant{Currency: ds.FrancsSuisse, Cent: 200}
	_, err = part.Participant.Update(ct.db)
	tu.AssertNoErr(t, err)

	bilan, err := logic.LoadDossiersFinance(ct.db, dossier1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, !bilan.Bilan().IsAcquitte()) // default camp prix is 100€

	err = ct.soldeDossier(dossier1.Id)
	tu.AssertNoErr(t, err)
	bilan, err = logic.LoadDossiersFinance(ct.db, dossier1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, bilan.Bilan().IsAcquitte() && bilan.Bilan().ApresPaiement().Cent == 0)
}
