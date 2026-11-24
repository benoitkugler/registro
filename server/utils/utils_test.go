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

func TestBuildURL(t *testing.T) {
	tu.Assert(t, BuildUrl("localhost:1323", "test", QPInt("key", 34)) == "http://localhost:1323/test?key=34")
}
