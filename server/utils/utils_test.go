package utils

import (
	"testing"
	"time"

	tu "registro/utils/testutils"
)

func TestFormatTime(t *testing.T) {
	for _, test := range []struct {
		t        time.Time
		expected string
	}{
		{time.Time{}, ""},
		{time.Date(2000, time.January, 3, 1, 1, 12, 0, time.Local), "3 Janv. 2000 à 01h01"},
	} {
		tu.Assert(t, FormatTime(test.t) == test.expected)
	}
}
