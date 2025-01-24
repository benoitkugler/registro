package search

import (
	"fmt"
	"testing"
	"time"

	pr "registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestRechercheRapide(t *testing.T) {
	db, err := tu.SampleDB.ConnectPostgres()
	tu.AssertNoErr(t, err)
	defer db.Close()

	m, err := pr.SelectAllPersonnes(db)
	tu.AssertNoErr(t, err)
	personnes := utils.MapValues(m)

	s := time.Now()
	res := Filter(personnes, "benoît kug")
	fmt.Println("filtered personnes in", time.Since(s))
	tu.Assert(t, len(res) > 0)

	tu.Assert(t, len(Filter(personnes, "*")) == len(personnes))
	tu.Assert(t, len(Filter(personnes, "")) == len(personnes))

	// s = time.Now()
	// fmt.Println("nb res :", len(base.RechercheRapideCamps("C2")))
	// fmt.Println("Camps :", time.Since(s))
	// if len(base.RechercheRapideCamps("*")) != len(base.Camps) {
	// 	t.FailNow()
	// }

	// s = time.Now()
	// fmt.Println("nb res :", len(base.RechercheRapideParticipants("C1")))
	// fmt.Println("Participants :", time.Since(s))
	// if len(base.RechercheRapideParticipants("*")) != len(base.Participants) {
	// 	t.FailNow()
	// }

	// s = time.Now()
	// fmt.Println("nb res:", len(base.RechercheRapideStructureaides("dr")))
	// fmt.Println("Structure aides :", time.Since(s))
	// if len(base.RechercheRapideStructureaides("*")) != len(base.Structureaides) {
	// 	t.FailNow()
	// }

	// s = time.Now()
	// fmt.Println("nb res:", len(base.RechercheRapideFactures("2018")))
	// fmt.Println(time.Since(s))
	// if len(base.RechercheRapideFactures("*")) != len(base.Factures) {
	// 	t.FailNow()
	// }
}
