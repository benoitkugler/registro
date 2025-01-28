package search

import (
	"testing"
	"time"

	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestSimilaires(t *testing.T) {
	entrant := pr.Etatcivil{
		Nom:               "NNS",
		Prenom:            "ben",
		Tels:              pr.Tels{"0675784512"},
		Sexe:              pr.Man,
		DateNaissance:     shared.NewDateFrom(time.Now()),
		Approfondissement: pr.ACanoe,
		Etudiant:          true,
	}
	existant := pr.Etatcivil{
		Nom:           "llmmlsd",
		Prenom:        "Bèn",
		Tels:          pr.Tels{"06-75-78-45-12", "0478458956"},
		Sexe:          pr.Woman,
		DateNaissance: shared.NewDateFrom(time.Now()),
		Fonctionnaire: true,
	}
	merged, conficts := Merge(entrant, existant)
	tu.Assert(t, conficts.Nom)
	tu.Assert(t, merged.Nom == "NNS")

	tu.Assert(t, !conficts.Prenom)
	tu.Assert(t, merged.Prenom == "ben")

	tu.Assert(t, conficts.Tels)

	tu.Assert(t, conficts.Sexe)
	tu.Assert(t, merged.Sexe == pr.Man)
	tu.Assert(t, !conficts.DateNaissance)
}
