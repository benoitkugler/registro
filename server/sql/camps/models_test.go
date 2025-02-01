package camps

import (
	"testing"
	"time"

	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	defautTaux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)

	camp1 := randCamp()
	camp1.IdTaux = defautTaux.Id
	camp1, err = camp1.Insert(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, camp1.Id != 0)
	personne, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := dossiers.Dossier{IdResponsable: personne.Id, IdTaux: defautTaux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	t.Run("equipiers", func(t *testing.T) {
		_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp1.Id, Roles: Roles{Direction, AutreRole}}.Insert(db)
		tu.AssertNoErr(t, err)
		// only one directeur
		_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp1.Id, Roles: Roles{Direction, Menage}}.Insert(db)
		tu.AssertErr(t, err)
	})

	t.Run("participants et groupes", func(t *testing.T) {
		part1 := randParticipant()
		part1.IdCamp, part1.IdPersonne, part1.IdDossier = camp1.Id, personne.Id, dossier.Id
		part1.IdTaux = camp1.IdTaux
		part1, err = part1.Insert(db)
		tu.AssertNoErr(t, err)

		groupe1, err := Groupe{IdCamp: camp1.Id, Nom: "1"}.Insert(db)
		tu.AssertNoErr(t, err)
		groupe2, err := Groupe{IdCamp: camp1.Id, Nom: "2"}.Insert(db)
		tu.AssertNoErr(t, err)

		err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: camp1.Id}).Insert(db)
		tu.AssertNoErr(t, err)

		err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: camp1.Id}).Insert(db)
		tu.AssertErr(t, err) // unicité
		err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe2.Id, IdCamp: camp1.Id}).Insert(db)
		tu.AssertErr(t, err) // unicité du participant

		err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: 0}).Insert(db)
		tu.AssertErr(t, err)

		_, err = DeleteGroupeById(db, groupe1.Id)
		tu.AssertNoErr(t, err)
		_, err = DeleteGroupeById(db, groupe2.Id)
		tu.AssertNoErr(t, err)
		_, err = DeleteParticipantById(db, part1.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("dossiers et taux", func(t *testing.T) {
		camp2 := randCamp()
		camp2.IdTaux = defautTaux.Id
		camp2, err := camp2.Insert(db)
		tu.AssertNoErr(t, err)

		taux, err := dossiers.Taux{}.Insert(db)
		tu.AssertNoErr(t, err)

		part1 := randParticipant()
		part1.IdCamp, part1.IdPersonne, part1.IdDossier = camp1.Id, personne.Id, dossier.Id

		part1.IdTaux = taux.Id
		_, err = part1.Insert(db)
		tu.AssertErr(t, err) // IdTaux n'est pas cohérent

		part1.IdTaux = defautTaux.Id
		part1, err = part1.Insert(db)
		tu.AssertNoErr(t, err)

		camp1.IdTaux = taux.Id
		_, err = camp1.Update(db)
		tu.AssertErr(t, err) // IdTaux n'est pas cohérent

		// deux camps sans taux sont OK
		part2 := part1
		part2.IdCamp = camp2.Id
		part2, err = part2.Insert(db)
		tu.AssertNoErr(t, err)

		camp3 := randCamp()
		camp3.IdTaux = taux.Id
		camp3, err = camp3.Insert(db)
		tu.AssertNoErr(t, err)

		part3 := randParticipant()
		part3.IdCamp, part3.IdPersonne, part3.IdDossier = camp3.Id, personne.Id, dossier.Id
		_, err = part3.Insert(db)
		tu.AssertErr(t, err) // IdTaux n'est pas cohérent
	})
}

func TestCamp_DateFin(t *testing.T) {
	for _, test := range []struct {
		debut    shared.Date
		duree    int
		expected shared.Date
	}{
		{shared.NewDate(2000, time.January, 15), 1, shared.NewDate(2000, time.January, 15)},
		{shared.NewDate(2000, time.January, 15), 10, shared.NewDate(2000, time.January, 24)},
		{shared.NewDate(2000, time.January, 30), 3, shared.NewDate(2000, time.February, 1)},
		{shared.NewDate(2000, time.December, 30), 3, shared.NewDate(2001, time.January, 1)},
	} {
		got := (&Camp{DateDebut: test.debut, Duree: test.duree}).DateFin()
		tu.Assert(t, got == test.expected)
	}
}
