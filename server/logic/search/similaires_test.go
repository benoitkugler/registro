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
	c1 := pr.Etatcivil{
		Nom:           "Benoit",
		Prenom:        "Kugler",
		Sexe:          pr.Man,
		DateNaissance: d1,
	}
	c2 := pr.Etatcivil{
		Nom:           "Léo",
		Prenom:        "Kugler",
		Sexe:          pr.Man,
		DateNaissance: d1,
	}
	c3 := pr.Etatcivil{
		Nom:           "Léa",
		Prenom:        "Kugler",
		Sexe:          pr.Woman,
		DateNaissance: d1,
	}
	c4 := pr.Etatcivil{
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
		fields    fields
		candidate pr.Etatcivil
		want      bool
	}{
		{fields{"Benoit", "Kugler", pr.Man, d1}, c1, true},
		{fields{"Benoit", "Kugler", pr.Man, d2}, c1, false},
		{fields{"Benoît", "Kugler", pr.Man, d1}, c1, true},
		{fields{"Benoît", "Kugler", pr.Man, d1}, c2, false},
		{fields{"Léo", "Kugler", pr.Woman, d1}, c3, false},
		{fields{"Dominique", "Kugler", pr.Woman, d1}, c4, true},
		{fields{"Dominique", "Kugler", pr.Man, d1}, c4, false},
	}
	for _, tt := range tests {
		ps := &PatternsSimilarite{
			Nom:           tt.fields.Nom,
			Prenom:        tt.fields.Prenom,
			Sexe:          tt.fields.Sexe,
			DateNaissance: tt.fields.DateNaissance,
		}
		ps.normalize()
		tu.Assert(t, ps.match(tt.candidate) == tt.want)
	}
}
