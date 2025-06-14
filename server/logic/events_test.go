package logic

import (
	"slices"
	"testing"

	tu "registro/utils/testutils"
)

func TestIterEvents(t *testing.T) {
	l := Events{
		{Content: Supprime{}},
		{Content: Supprime{}},
		{Content: Sondage{}},
		{Content: Attestation{}},
		{Content: Supprime{}},
		{Content: Facture{}},
		{Content: Supprime{}},
		{Content: Facture{}},
	}
	tu.Assert(t, len(slices.Collect(IterContentBy[Supprime](l))) == 4)
	tu.Assert(t, len(slices.Collect(IterContentBy[Message](l))) == 0)
}
