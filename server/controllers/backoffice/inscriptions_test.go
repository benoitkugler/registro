package backoffice

import (
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestController_getInscriptions(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := personnes.Personne{}.Insert(db)
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

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	out, err := ct.getInscriptions()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out) == 2)
	tu.Assert(t, out[0].Dossier.Id == dossier1.Id)
}
