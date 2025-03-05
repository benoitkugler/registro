package backoffice

import (
	"fmt"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestOffuscateur(t *testing.T) {
	offuscateur := newOffuscateur[int64]("VI", 8, 3)
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

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, fs.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	out, err := ct.searchDossiers(SearchDossierIn{Pattern: OffuscateurVirements.Mask(dossier1.Id)})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 1)

	out, err = ct.searchDossiers(SearchDossierIn{Pattern: "test"})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Dossiers) == 0)
}

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

	structure, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, fs.NewFileSystem(t.TempDir()), config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	part, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp1.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)

	part.Statut = cps.Inscrit
	part.QuotientFamilial = 48
	err = ct.updateParticipant(part)
	tu.AssertNoErr(t, err)

	aide, err := ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)
	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test.png")
	tu.AssertNoErr(t, err)

	files, err := fs.SelectAllFiles(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 1)

	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp2.Id, IdPersonne: pe1.Id})
	tu.AssertErr(t, err) // inconsistent taux

	err = ct.deleteParticipant(part.Id)
	tu.AssertNoErr(t, err)

	files, err = fs.SelectAllFiles(ct.db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 0)

	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier1.Id, IdCamp: camp2.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err) // now the change of taux is OK
}

const pngData = "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A\x00\x00\x00\x0D\x49\x48\x44\x52" +
	"\x00\x00\x01\x00\x00\x00\x01\x00\x01\x03\x00\x00\x00\x66\xBC\x3A" +
	"\x25\x00\x00\x00\x03\x50\x4C\x54\x45\xB5\xD0\xD0\x63\x04\x16\xEA" +
	"\x00\x00\x00\x1F\x49\x44\x41\x54\x68\x81\xED\xC1\x01\x0D\x00\x00" +
	"\x00\xC2\xA0\xF7\x4F\x6D\x0E\x37\xA0\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\xBE\x0D\x21\x00\x00\x01\x9A\x60\xE1\xD5\x00\x00\x00\x00\x49" +
	"\x45\x4E\x44\xAE\x42\x60\x82"

func TestController_aides(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1, Places: 20, AgeMin: 6, AgeMax: 12}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	part, err := cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	structure, err := cps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, fs.NewFileSystem(t.TempDir()), config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	aide, err := ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)

	aide.Valide = true
	aide.Valeur = ds.NewEuros(26)
	aide.ParJour = true
	err = ct.updateAide(aide)
	tu.AssertNoErr(t, err)

	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test1.png")
	tu.AssertNoErr(t, err)

	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test2.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteAideJustificatif(aide.Id)
	tu.AssertNoErr(t, err)

	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test3.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteAide(aide.Id)
	tu.AssertNoErr(t, err)

	aide, err = ct.createAide(AidesCreateIn{IdParticipant: part.Id, IdStructure: structure.Id})
	tu.AssertNoErr(t, err)

	err = ct.uploadAideJustificatif(aide.Id, []byte(pngData), "test3.png")
	tu.AssertNoErr(t, err)

	err = ct.deleteDossier(dossier1.Id)
	tu.AssertNoErr(t, err)
}

func TestController_paiements(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false, Etatcivil: pr.Etatcivil{DateNaissance: shared.Date(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, fs.NewFileSystem(t.TempDir()), config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	out, err := ct.createPaiement(dossier1.Id)
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
