package matching

import (
	"encoding/json"
	"os"
	"testing"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func loadPersonnes(t *testing.T) []pr.Personne {
	b, err := os.ReadFile("test/samples.json")
	tu.AssertNoErr(t, err)
	var personnes []pr.Personne
	err = json.Unmarshal(b, &personnes)
	tu.AssertNoErr(t, err)
	return personnes
}

func TestLoadPattern(t *testing.T) {
	db := tu.NewTestDB(t, "../../sql/personnes/gen_create.sql")
	defer db.Remove()

	personnes := loadPersonnes(t)
	for _, p := range personnes {
		_, err := p.Insert(db)
		tu.AssertNoErr(t, err)
	}

	pers, err := SelectAllPatternSimilaires(db)
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
	tu.Assert(t, len(res1) > len(res2))

	_, res3 := ChercheSimilaires(personnes, PatternsSimilarite{
		Mail: personnes[0].Mail,
	})
	tu.Assert(t, len(res3) == 1) // all mails are distincts
}
