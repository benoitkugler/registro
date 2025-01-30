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

	camp, err := randCamp().Insert(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, camp.Id != 0)
	personne, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp.Id, Roles: Roles{Direction, AutreRole}}.Insert(db)
	tu.AssertNoErr(t, err)
	// only one directeur
	_, err = Equipier{IdPersonne: personne.Id, IdCamp: camp.Id, Roles: Roles{Direction, Menage}}.Insert(db)
	tu.Assert(t, err != nil)

	// participants et groupe
	dossier, err := dossiers.Dossier{IdResponsable: personne.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	part1 := randParticipant()
	part1.IdCamp, part1.IdPersonne, part1.IdDossier = camp.Id, personne.Id, dossier.Id
	part1, err = part1.Insert(db)
	tu.AssertNoErr(t, err)

	groupe1, err := Groupe{IdCamp: camp.Id, Nom: "1"}.Insert(db)
	tu.AssertNoErr(t, err)
	groupe2, err := Groupe{IdCamp: camp.Id, Nom: "2"}.Insert(db)
	tu.AssertNoErr(t, err)

	err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: camp.Id}).Insert(db)
	tu.AssertNoErr(t, err)

	err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: camp.Id}).Insert(db)
	tu.Assert(t, err != nil) // unicité
	err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe2.Id, IdCamp: camp.Id}).Insert(db)
	tu.Assert(t, err != nil) // unicité du participant

	err = GroupeParticipant(GroupeParticipant{IdParticipant: part1.Id, IdGroupe: groupe1.Id, IdCamp: 0}).Insert(db)
	tu.Assert(t, err != nil)
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
