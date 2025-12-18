package utils

import (
	"testing"

	tu "registro/utils/testutils"
)

func TestRandColor(t *testing.T) {
	for range 200 {
		tu.Assert(t, len(RandColor()) == 7)
	}
}
