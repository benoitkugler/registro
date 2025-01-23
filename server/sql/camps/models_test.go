package camps

import (
	"testing"
	"time"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestCamp_DateFin(t *testing.T) {
	for _, test := range []struct {
		debut    pr.Date
		duree    int
		expected pr.Date
	}{
		{pr.NewDate(2000, time.January, 15), 1, pr.NewDate(2000, time.January, 15)},
		{pr.NewDate(2000, time.January, 15), 10, pr.NewDate(2000, time.January, 24)},
		{pr.NewDate(2000, time.January, 30), 3, pr.NewDate(2000, time.February, 1)},
		{pr.NewDate(2000, time.December, 30), 3, pr.NewDate(2001, time.January, 1)},
	} {
		got := (&Camp{DateDebut: test.debut, Duree: test.duree}).DateFin()
		tu.Assert(t, got == test.expected)
	}
}
