package search

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func loadPersonnes(t *testing.T) pr.Personnes {
	b, err := os.ReadFile("test/samples.json")
	tu.AssertNoErr(t, err)
	var personnes []pr.Personne
	err = json.Unmarshal(b, &personnes)
	tu.AssertNoErr(t, err)
	out := make(pr.Personnes)
	for _, personne := range personnes {
		out[personne.Id] = personne
	}
	return out
}

func TestLoadPattern(t *testing.T) {
	db := tu.NewTestDB(t, "../../sql/personnes/gen_create.sql")
	defer db.Remove()

	personnes := loadPersonnes(t)
	for _, p := range personnes {
		_, err := p.Insert(db)
		tu.AssertNoErr(t, err)
	}

	pers, err := SelectAllFieldsForSimilaires(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(pers) == len(personnes))
	tu.Assert(t, pers[0].Adresse == "")
}

func TestLoadAndSearchSimilaires(t *testing.T) {
	personnes := loadPersonnes(t)

	sm1, res1 := ChercheSimilaires(personnes, PatternsSimilarite{
		Nom:    "kug",
		Prenom: "dede",
	})
	tu.Assert(t, sm1 == 2+3)
	tu.Assert(t, len(res1) > 0)

	_, res2 := ChercheSimilaires(personnes, PatternsSimilarite{
		Nom:    "Kug ",
		Prenom: "dédé ",
	})
	tu.Assert(t, len(res1) == len(res2))

	_, ok := Match(personnes, PatternsSimilarite{
		Nom:           "JOnac",
		Prenom:        "Tom",
		DateNaissance: shared.NewDate(2021, time.May, 6),
		Sexe:          pr.Woman,
	})
	tu.Assert(t, ok)
}

func Test_normalize(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"Benoit", "benoit"},
		{"Benoît", "benoit"},
		{"Benoît!ï", "benoiti"},
		{"Jean-Pierre", "jeanpierre"},
		{"Jean-Pier re", "jeanpierre"},
	}
	for _, tt := range tests {
		tu.Assert(t, Normalize(tt.args) == tt.want)
	}
}

func TestPatternsSimilarite_match(t *testing.T) {
	d1 := shared.NewDate(1994, time.April, 2)
	d2 := shared.NewDate(1994, time.April, 3)
	c1 := pr.Identite{
		Nom:           "Benoit",
		Prenom:        "Kugler",
		Sexe:          pr.Man,
		DateNaissance: d1,
	}
	c2 := pr.Identite{
		Nom:           "Léo",
		Prenom:        "Kugler",
		Sexe:          pr.Man,
		DateNaissance: d1,
	}
	c3 := pr.Identite{
		Nom:           "Léa",
		Prenom:        "Kugler",
		Sexe:          pr.Woman,
		DateNaissance: d1,
	}
	c4 := pr.Identite{
		Nom:           "Dominique",
		Prenom:        "Kugler",
		Sexe:          pr.Woman,
		DateNaissance: d1,
	}
	type fields struct {
		Nom           string
		Prenom        string
		Sexe          pr.Sexe
		DateNaissance shared.Date
	}

	tests := []struct {
		fields    PatternsSimilarite
		candidate pr.Identite
		want      bool
	}{
		{PatternsSimilarite{"Benoit", "Kugler", d1, pr.Man}, c1, true},
		{PatternsSimilarite{"Benoit", "Kugler", d2, pr.Man}, c1, false},
		{PatternsSimilarite{"Benoît", "Kugler", d1, pr.Man}, c1, true},
		{PatternsSimilarite{"Benoît", "Kugler", d1, pr.Man}, c2, false},
		{PatternsSimilarite{"Léo", "Kugler", d1, pr.Woman}, c3, false},
		{PatternsSimilarite{"Dominique", "Kugler", d1, pr.Woman}, c4, true},
		{PatternsSimilarite{"Dominique", "Kugler", d1, pr.Man}, c4, false},
		{PatternsSimilarite{"Dominique", "Kugler", d1, pr.Man}, c4, false},
		{PatternsSimilarite{"Dominique", "Kugler", d1, pr.NoSexe}, c4, true}, // ignore sexe
	}
	for _, tt := range tests {
		ps := tt.fields
		ps.normalize()
		tu.Assert(t, ps.match(tt.candidate) == tt.want)
	}
}
