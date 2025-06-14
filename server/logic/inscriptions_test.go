package logic

import (
	"reflect"
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestIdentifieProfil(t *testing.T) {
	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{IsTemp: false}.Insert(db)
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

func TestStatutBypassRights_resolve(t *testing.T) {
	directorRights := StatutBypassRights{false, true, false}
	backofficeRights := StatutBypassRights{true, true, true}
	tests := []struct {
		fields StatutBypassRights
		args   cps.StatutCauses
		want   StatutExt
	}{
		// problème d'age
		{directorRights, cps.StatutCauses{AgeMin: false}, StatutExt{AllowedChanges: nil, Validable: false}},
		// problème de place -> OK
		{directorRights, cps.StatutCauses{AgeMin: true, AgeMax: true}, StatutExt{AllowedChanges: []cps.StatutParticipant{cps.Inscrit}, Validable: true}},
		// valide
		{directorRights, cps.StatutCauses{AgeMin: true, AgeMax: true, EquilibreGF: true, Place: true}, StatutExt{AllowedChanges: nil, Validable: true}},

		// problème d'age
		{backofficeRights, cps.StatutCauses{AgeMin: false}, StatutExt{AllowedChanges: []cps.StatutParticipant{cps.Inscrit}, Validable: true}},
		// problème de place -> OK
		{backofficeRights, cps.StatutCauses{AgeMin: true, AgeMax: true}, StatutExt{AllowedChanges: []cps.StatutParticipant{cps.Inscrit}, Validable: true}},
		// valide
		{backofficeRights, cps.StatutCauses{AgeMin: true, AgeMax: true, EquilibreGF: true, Place: true}, StatutExt{AllowedChanges: []cps.StatutParticipant{cps.AttenteProfilInvalide, cps.AttenteCampComplet}, Validable: true}},
	}
	for _, tt := range tests {
		got := tt.fields.resolve(tt.args, cps.AStatuer)
		tu.Assert(t, reflect.DeepEqual(got.AllowedChanges, tt.want.AllowedChanges))
		tu.Assert(t, got.Validable == tt.want.Validable)
		tu.Assert(t, got.IsAllowed(got.Statut))
	}

	tu.Assert(t, !backofficeRights.resolve(cps.StatutCauses{}, cps.Inscrit).Validable)
}
