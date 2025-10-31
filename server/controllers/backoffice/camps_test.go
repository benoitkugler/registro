package backoffice

import (
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cp "registro/sql/camps"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestCRUD(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, "", "", files.FileSystem{}, config.SMTP{}, config.Asso{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	t.Run("camps", func(t *testing.T) {
		camps, err := ct.getCamps("localhost")
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(camps) == 0)

		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)
		tu.Assert(t, camp.Camp.Camp.Statut == cps.VisibleFerme)

		camps, err = ct.getCamps("localhost")
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(camps) == 1)

		camp.Camp.Camp.Statut = cps.Ouvert
		camp.Camp.Camp.Navette = cps.OptionNavette{Actif: true, Commentaire: "7 € Aller retour"}
		updated, err := ct.updateCamp(camp.Camp.Camp)
		tu.AssertNoErr(t, err)
		tu.Assert(t, updated.Camp.Statut == cps.Ouvert)

		err = ct.deleteCamp(camp.Camp.Camp.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("equipiers", func(t *testing.T) {
		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)

		_, err = ct.createEquipier(CreateEquipierIn{pe.Id, camp.Camp.Camp.Id, cps.Roles{cp.Direction}})
		tu.AssertNoErr(t, err)
	})

	t.Run("documents", func(t *testing.T) {
		camp, err := ct.createCamp()
		tu.AssertNoErr(t, err)

		_, err = ct.getCampDocument(camp.Camp.Camp.Id)
		tu.AssertNoErr(t, err)
	})
}

func Test_lastTaux(t *testing.T) {
	tests := []struct {
		camps cp.Camps
		want  ds.IdTaux
	}{
		{nil, 1},
		{cps.Camps{2: {IdTaux: 2, DateDebut: shared.NewDate(2000, 1, 1)}}, 2},
		{cps.Camps{
			2: {IdTaux: 2, DateDebut: shared.NewDate(2000, 1, 1)},
			3: {IdTaux: 3, DateDebut: shared.NewDate(2001, 1, 1)},
		}, 3},
	}
	for _, tt := range tests {
		tu.Assert(t, lastTaux(tt.camps) == tt.want)
	}
}

func TestExportParticipants(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	ct, err := NewController(db.DB, crypto.Encrypter{}, "", "", files.FileSystem{}, config.SMTP{}, config.Asso{}, config.Joomeo{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	resp, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "REspo", Prenom: "Sable"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe1, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "REspo", Prenom: "Huge"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: "Autre", Prenom: "Hugette"}}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1, err := ct.createCamp()
	tu.AssertNoErr(t, err)
	camp2, err := ct.createCamp()
	tu.AssertNoErr(t, err)

	dossier, err := ct.createDossier(resp.Id)
	tu.AssertNoErr(t, err)
	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier.Id, IdCamp: camp1.Camp.Camp.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)
	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier.Id, IdCamp: camp1.Camp.Camp.Id, IdPersonne: pe2.Id})
	tu.AssertNoErr(t, err)
	_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier.Id, IdCamp: camp2.Camp.Camp.Id, IdPersonne: pe1.Id})
	tu.AssertNoErr(t, err)

	paiement1, err := ct.createPaiement(true, dossier.Id)
	tu.AssertNoErr(t, err)
	paiement1.Montant.Cent = 4568
	_, err = paiement1.Update(ct.db)
	tu.AssertNoErr(t, err)

	paiement2, err := ct.createPaiement(false, dossier.Id)
	tu.AssertNoErr(t, err)
	paiement2.Montant.Cent = 4335
	_, err = paiement2.Update(ct.db)
	tu.AssertNoErr(t, err)

	content, name, err := ct.exportListeParticipants(time.Now().Year())
	tu.AssertNoErr(t, err)
	tu.Write(t, name, content)
}
