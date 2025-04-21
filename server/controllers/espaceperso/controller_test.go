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

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{}, files.NewFileSystem(os.TempDir()))

	err = ct.createAide(dossier.Id, camps.Aide{IdStructureaide: st.Id, IdParticipant: pa.Id, Valeur: ds.NewEuros(456.4)}, tu.PngData, "test.png")
	tu.AssertNoErr(t, err)
}
