package dossiers

import (
	"testing"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestDossierKeyV1Uniqueness(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	respo, err := pr.Personne{}.Insert(db.DB)
	tu.AssertNoErr(t, err)
	taux, err := Taux{Label: "EUR", Euros: 1000}.Insert(db.DB)
	tu.AssertNoErr(t, err)

	insert := func(key string) Dossier {
		t.Helper()
		dossier, err := Dossier{IdTaux: taux.Id, IdResponsable: respo.Id, KeyV1: key}.Insert(db.DB)
		tu.AssertNoErr(t, err)
		return dossier
	}

	firstEmpty := insert("")
	secondEmpty := insert("")
	tu.Assert(t, firstEmpty.Id != secondEmpty.Id)

	const key = "legacy-key"
	withKey := insert(key)
	_, err = Dossier{IdTaux: taux.Id, IdResponsable: respo.Id, KeyV1: key}.Insert(db.DB)
	tu.AssertErr(t, err)

	got, found, err := SelectDossierByKeyV1(db.DB, key)
	tu.AssertNoErr(t, err)
	tu.Assert(t, found && got.Id == withKey.Id)
}
