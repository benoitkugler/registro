package espaceperso

import (
	"os"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/generators/pdfcreator"
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

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(os.TempDir()), config.Joomeo{})

	err = ct.createAide(dossier.Id, cps.Aide{IdStructureaide: st.Id, IdParticipant: pa.Id, Valeur: ds.NewEuros(456.4)}, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)
}

func TestJoomeo(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	// this email is used on the DEV joomeo account
	pe, err := pr.Personne{Etatcivil: pr.Etatcivil{Mail: "x.ben.x@free.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)

	pe2, err := pr.Personne{Etatcivil: pr.Etatcivil{Mail: "xxxxx@free.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: pe.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier2, err := ds.Dossier{IdTaux: 1, IdResponsable: pe2.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	tu.LoadEnv(t, "../../env.sh")
	joomeo, err := config.NewJoomeo()
	tu.AssertNoErr(t, err)
	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, fs.NewFileSystem(os.TempDir()), joomeo)

	data, err := ct.loadJoomeo(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.Loggin != "" && data.Password != "")

	data, err = ct.loadJoomeo(dossier2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.Loggin == "" && data.Password == "")
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
	fs, err := ct.loadFichesanitaires(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(fs.Fiches) == 1)
	tu.Assert(t, fs.Fiches[0].Fichesanitaire.IdPersonne == mineur.Id)
	tu.Assert(t, !fs.Fiches[0].IsLocked && fs.Fiches[0].State == pr.NoFiche)
	tu.Assert(t, len(fs.Fiches[0].VaccinsFiles) == 1)
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

func TestDownloadDocuments(t *testing.T) {
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
}
