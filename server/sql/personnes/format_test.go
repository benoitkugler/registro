package personnes

import (
	"testing"

	tu "registro/utils/testutils"
)

func TestTels(t *testing.T) {
	tu.Assert(t, Tels{}.String() == "")
	tu.Assert(t, Tels{"", "8"}.String() == "8")
	tu.Assert(t, Tels{"7", "8"}.String() == "7;8")
}
