package backoffice

import (
	"database/sql"
	"testing"
	"time"

	"registro/config"
	"registro/controllers/logic"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
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

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	out, err := ct.getInscriptions()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out) == 2)
	tu.Assert(t, out[0].Dossier.Id == dossier1.Id)
}

func TestController_searchSimilaires(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	err := utils.InTx(db.DB, func(tx *sql.Tx) error {
		for range [2000]int{} {
			_, err := pr.Personne{}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, err = ct.searchSimilaires(1)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Since(ti) < 50*time.Millisecond)
}

func TestIdentifieProfil(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: true}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{IsTemp: false}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier1, err := ds.Dossier{IdResponsable: pe1.Id, IdTaux: 1, MomentInscription: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp1.Id, IdPersonne: pe1.Id, IdDossier: dossier1.Id, IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	fi, err := files.File{}.Insert(db)
	tu.AssertNoErr(t, err)
	demande, err := files.Demande{MaxDocs: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	err = files.FilePersonne{IdFile: fi.Id, IdPersonne: pe1.Id, IdDemande: demande.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = IdentifiePersonne(db.DB, IdentTarget{
		IdTemporaire: pe1.Id,
		Rattache:     true,
		RattacheTo:   pe2.Id,
	})
	tu.AssertNoErr(t, err)

	links, err := files.SelectFilePersonnesByIdPersonnes(db, pe2.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(links) == 1)
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

	ct, err := NewController(db.DB, crypto.Encrypter{}, files.FileSystem{}, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	err = ct.valideInscription(dossier1.Id)
	tu.AssertNoErr(t, err)

	ld, err := logic.LoadDossiers(db, dossier1.Id)
	tu.AssertNoErr(t, err)
	data := ld.For(dossier1.Id)
	for _, part := range data.Participants {
		tu.Assert(t, part.Statut == cps.AttenteProfilInvalide)
	}
	tu.Assert(t, data.Dossier.IsValidated)
}
