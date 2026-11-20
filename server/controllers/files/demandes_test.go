package files

import (
	"testing"

	"registro/config"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"
)

// TestLoadDemandes checks the two-phase resolution: candidates are first
// resolved from the Asso settings and the camp, then filtered per participant.
func TestLoadDemandes(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1, DateDebut: sh.NewDate(2026, 7, 1)}.Insert(db)
	tu.AssertNoErr(t, err)

	// camp-specific demandes
	customBoth, err := fs.Demande{
		Categorie: fs.NoBuiltin, Description: "Attestation", MaxDocs: 1, Nationnalite: fs.Both,
	}.Insert(db)
	tu.AssertNoErr(t, err)
	customCH, err := fs.Demande{
		Categorie: fs.NoBuiltin, Description: "Attestation Suisse", MaxDocs: 1, Nationnalite: fs.OnlyCh,
	}.Insert(db)
	tu.AssertNoErr(t, err)
	for _, custom := range []fs.Demande{customBoth, customCH} {
		tu.AssertNoErr(t, fs.DemandeCamp{IdCamp: camp.Id, IdDemande: custom.Id}.Insert(db))
	}

	vaccin, carteId, err := DemandesVaccinEtCarteId(db)
	tu.AssertNoErr(t, err)

	t.Run("camp columns depend on Asso settings", func(t *testing.T) {
		loader, err := LoadDemandes(db, config.AskedDocs{}, camp.Id)
		tu.AssertNoErr(t, err)
		cols := loader.DemandesForCamp(camp.Id)
		tu.Assert(t, len(cols) == 3) // 2 custom + vaccin
		_, hasVaccin := cols[vaccin.Id]
		tu.Assert(t, hasVaccin)
		_, hasCarteId := cols[carteId.Id]
		tu.Assert(t, !hasCarteId)

		loader, err = LoadDemandes(db, config.AskedDocs{CarteId: true}, camp.Id)
		tu.AssertNoErr(t, err)
		cols = loader.DemandesForCamp(camp.Id)
		tu.Assert(t, len(cols) == 4)
		_, hasCarteId = cols[carteId.Id]
		tu.Assert(t, hasCarteId)
	})

	// 14 at camp start
	mineur, err := pr.Personne{Identite: pr.Identite{DateNaissance: sh.NewDate(2012, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)
	// 36 at camp start
	majeur, err := pr.Personne{Identite: pr.Identite{DateNaissance: sh.NewDate(1990, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)
	// 36 at camp start, Swiss
	majeurCH, err := pr.Personne{Identite: pr.Identite{
		DateNaissance: sh.NewDate(1990, 1, 1), Nationnalite: pr.Nationnalite{IsSuisse: true},
	}}.Insert(db)
	tu.AssertNoErr(t, err)
	temp, err := pr.Personne{IsTemp: true, Identite: pr.Identite{DateNaissance: sh.NewDate(2012, 1, 1)}}.Insert(db)
	tu.AssertNoErr(t, err)

	forParticipant := func(loader DemandesLoader, personne pr.Personne, statut cps.StatutParticipant) DemandesParticipant {
		return loader.ForParticipant(ParticipantDemandesArgs{
			IdCamp: camp.Id, AgeCamp: camp, Personne: personne, Statut: statut,
		})
	}

	t.Run("age drives fiche sanitaire and vaccins", func(t *testing.T) {
		loader, err := LoadDemandes(db, config.AskedDocs{}, camp.Id)
		tu.AssertNoErr(t, err)

		got := forParticipant(loader, mineur, cps.Inscrit)
		tu.Assert(t, got.FicheSanitaire)
		tu.Assert(t, got.Has(vaccin.Id))
		tu.Assert(t, got.Has(customBoth.Id))

		got = forParticipant(loader, majeur, cps.Inscrit)
		tu.Assert(t, !got.FicheSanitaire)
		tu.Assert(t, !got.Has(vaccin.Id)) // vaccins dropped when no fiche sanitaire
		tu.Assert(t, got.Has(customBoth.Id))

		loader, err = LoadDemandes(db, config.AskedDocs{FicheSanitaireMajeurs: true}, camp.Id)
		tu.AssertNoErr(t, err)
		got = forParticipant(loader, majeur, cps.Inscrit)
		tu.Assert(t, got.FicheSanitaire)
		tu.Assert(t, got.Has(vaccin.Id))
	})

	t.Run("nationality filters camp demandes", func(t *testing.T) {
		loader, err := LoadDemandes(db, config.AskedDocs{}, camp.Id)
		tu.AssertNoErr(t, err)

		got := forParticipant(loader, majeur, cps.Inscrit)
		tu.Assert(t, !got.Has(customCH.Id))

		got = forParticipant(loader, majeurCH, cps.Inscrit)
		tu.Assert(t, got.Has(customCH.Id))
		tu.Assert(t, got.Has(customBoth.Id))
	})

	t.Run("temp and non-inscrit participants are excluded", func(t *testing.T) {
		loader, err := LoadDemandes(db, config.AskedDocs{}, camp.Id)
		tu.AssertNoErr(t, err)

		got := forParticipant(loader, temp, cps.Inscrit)
		tu.Assert(t, !got.FicheSanitaire && len(got.Demandes) == 0)

		got = forParticipant(loader, mineur, cps.AStatuer)
		tu.Assert(t, !got.FicheSanitaire && len(got.Demandes) == 0)
	})
}
