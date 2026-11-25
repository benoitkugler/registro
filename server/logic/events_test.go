package logic

import (
	"testing"
	"time"

	sh "registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestIterEvents(t *testing.T) {
	l := Events{
		{Content: SupprimeEvt{}},
		{Content: SupprimeEvt{}},
		{Content: SondageEvt{}},
		{Content: AttestationEvt{}},
		{Content: SupprimeEvt{}},
		{Content: DemandePaiementEvt{}},
		{Content: SupprimeEvt{}},
		{Content: DemandePaiementEvt{}},
	}
	tu.Assert(t, len(EventsBy[SupprimeEvt](l)) == 4)
	tu.Assert(t, len(EventsBy[MessageEvt](l)) == 0)
}

func TestHasDemandePaiement(t *testing.T) {
	first := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	last := time.Date(2026, time.February, 3, 10, 0, 0, 0, time.UTC)

	date, has := Events{}.HasDemandePaiement()
	tu.Assert(t, !has)
	tu.Assert(t, date.String() == "")

	l := Events{
		{Id: 1, Created: first, Content: DemandePaiementEvt{}},
		{Id: 2, Created: last.Add(24 * time.Hour), Content: SupprimeEvt{}},
		{Id: 3, Created: last, Content: DemandePaiementEvt{}},
	}
	date, has = l.HasDemandePaiement()
	tu.Assert(t, has)
	tu.Assert(t, date == sh.NewDateFrom(last))
}
