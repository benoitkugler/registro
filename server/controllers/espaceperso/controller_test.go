package espaceperso

import (
	"os"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/generators/pdfcreator"
	"registro/immich"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func Test_createAide(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	pa, err := cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	st, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(os.TempDir()), config.Immich{})

	err = ct.createAide(dossier.Id, cps.Aide{IdStructureaide: st.Id, IdParticipant: pa.Id, Valeur: ds.NewEuros(456.4)}, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)
}

func TestPhotos(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	pe2, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier2, err := ds.Dossier{IdTaux: 1, IdResponsable: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp1.Id, IdDossier: dossier1.Id, IdPersonne: pe.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp2.Id, IdDossier: dossier2.Id, IdPersonne: pe.Id, IdTaux: 1, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	tu.LoadEnv(t, "../../env.sh")
	photos, err := config.NewImmich()
	tu.AssertNoErr(t, err)
	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(os.TempDir()), photos)

	api := immich.NewApi(ct.immich)
	album, err := api.CreateAlbum("__TEST")
	tu.AssertNoErr(t, err)
	defer api.DeleteAlbum(album.Id)

	camp1.AlbumID = string(album.Id)
	_, err = camp1.Update(ct.db)
	tu.AssertNoErr(t, err)

	data, err := ct.loadPhotos(dossier1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(data) == 1)

	data, err = ct.loadPhotos(dossier2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(data) == 0)
}

func Test_loadFichesanitaires(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	fsys := fs.NewFileSystem(os.TempDir())

	now := time.Now()

	camp1, err := cps.Camp{IdTaux: 1, DateDebut: shared.Date(now), Duree: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := cps.Camp{IdTaux: 1, DateDebut: shared.Date(now), Duree: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	mineur, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(now.Add(-15 * 365 * 24 * time.Hour))}}.Insert(db)
	tu.AssertNoErr(t, err)
	mineur2, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(now.Add(-15 * 365 * 24 * time.Hour))}}.Insert(db)
	tu.AssertNoErr(t, err)
	majeur, err := pr.Personne{Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(now.Add(-30 * 365 * 24 * time.Hour))}}.Insert(db)
	tu.AssertNoErr(t, err)

	// vaccins
	file1, err := fs.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = fs.FilePersonne{IdFile: file1.Id, IdPersonne: mineur.Id, IdDemande: fs.IdDemande(fs.Vaccins)}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = fs.UploadFile(fsys, db, file1.Id, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: mineur.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{Statut: cps.Inscrit, IdCamp: camp1.Id, IdPersonne: mineur.Id, IdTaux: 1, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{Statut: cps.Inscrit, IdCamp: camp1.Id, IdPersonne: majeur.Id, IdTaux: 1, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{Statut: cps.Inscrit, IdCamp: camp2.Id, IdPersonne: mineur.Id, IdTaux: 1, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{Statut: cps.Inscrit, IdCamp: camp2.Id, IdPersonne: majeur.Id, IdTaux: 1, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{Statut: cps.AttenteCampComplet, IdCamp: camp2.Id, IdPersonne: mineur2.Id, IdTaux: 1, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}
	ext, err := logic.LoadDossier(ct.db, dossier.Id)
	tu.AssertNoErr(t, err)
	fiches, err := loadFichesanitaires(ct.db, ext)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(fiches) == 1)
	tu.Assert(t, fiches[0].Fichesanitaire.IdPersonne == mineur.Id)
	tu.Assert(t, !fiches[0].IsLocked && fiches[0].State == pr.NoFiche)
}

func Test_sondages(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}
	l, err := ct.loadSondages(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 0)

	ev, err := events.Event{IdDossier: dossier.Id, Kind: events.Sondage}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	err = events.EventSondage{IdEvent: ev.Id, IdCamp: camp.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	l, err = ct.loadSondages(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 1)

	sondage := l[0]
	sondage.Sondage.Ambiance = 3
	err = ct.updateSondage(dossier.Id, sondage.Sondage.Id, sondage.Sondage.IdCamp, sondage.Sondage.ReponseSondage)
	tu.AssertNoErr(t, err)
}

func TestDownloadJustificatifs(t *testing.T) {
	err := pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, Statut: cps.Inscrit}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	_, err = ct.renderAttestationPresence(dossier.Id)
	tu.AssertNoErr(t, err)

	_, err = ct.renderFacture(dossier.Id)
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

func TestPlaceLiberee(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	pa, err := cps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id, Statut: cps.AttenteCampComplet}.Insert(db)
	tu.AssertNoErr(t, err)

	ev, err := events.Event{IdDossier: dossier.Id, Kind: events.PlaceLiberee, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = events.EventPlaceLiberee{IdEvent: ev.Id, IdParticipant: pa.Id, Accepted: false}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, asso: asso, smtp: smtp}
	err = ct.acceptePlaceLiberee(dossier.Id, ev.Id)
	tu.AssertNoErr(t, err)

	err = ct.acceptePlaceLiberee(dossier.Id, ev.Id)
	tu.AssertErr(t, err) // already accepted
}
