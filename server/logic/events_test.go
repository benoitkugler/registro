package logic

import (
	"slices"
	"testing"

	tu "registro/utils/testutils"
)

func TestIterEvents(t *testing.T) {
	l := Events{
		{Content: SupprimeEvt{}},
		{Content: SupprimeEvt{}},
		{Content: SondageEvt{}},
		{Content: AttestationEvt{}},
		{Content: SupprimeEvt{}},
		{Content: FactureEvt{}},
		{Content: SupprimeEvt{}},
		{Content: FactureEvt{}},
	}
	tu.Assert(t, len(slices.Collect(IterEventsBy[SupprimeEvt](l))) == 4)
	tu.Assert(t, len(slices.Collect(IterEventsBy[MessageEvt](l))) == 0)
}
