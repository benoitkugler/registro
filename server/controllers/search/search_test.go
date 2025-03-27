package search

import (
	"fmt"
	"testing"
	"time"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestRechercheRapide(t *testing.T) {
	m := make(pr.Personnes)
	for _, p := range loadPersonnes(t) {
		m[p.Id] = p
	}

	s := time.Now()
	res := FilterPersonnes(m, "benoît kug")
	fmt.Println("filtered personnes in", time.Since(s))
	tu.Assert(t, len(res) > 0)

	tu.Assert(t, len(FilterPersonnes(m, "")) == len(m))
}
