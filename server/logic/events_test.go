package logic

import (
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
	tu.Assert(t, len(EventsBy[SupprimeEvt](l)) == 4)
	tu.Assert(t, len(EventsBy[MessageEvt](l)) == 0)
}
