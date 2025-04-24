package espaceperso

import (
	"os"
	"testing"

	"registro/config"
	"registro/crypto"
	"registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
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

	camp, err := camps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	pa, err := camps.Participant{IdTaux: 1, IdCamp: camp.Id, IdPersonne: pe.Id, IdDossier: dossier.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	st, err := camps.Structureaide{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, files.NewFileSystem(os.TempDir()), config.Joomeo{})

	err = ct.createAide(dossier.Id, camps.Aide{IdStructureaide: st.Id, IdParticipant: pa.Id, Valeur: ds.NewEuros(456.4)}, tu.PngData, "test.png")
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
	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, files.NewFileSystem(os.TempDir()), joomeo)

	data, err := ct.loadJoomeo(dossier.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.Loggin != "" && data.Password != "")

	data, err = ct.loadJoomeo(dossier2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.Loggin == "" && data.Password == "")
}
