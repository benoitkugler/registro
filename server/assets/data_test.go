package assets

import (
	"testing"

	tu "registro/utils/testutils"
)

func Test(t *testing.T) {
	tu.Assert(t, len(DefaultMiniaturePNG) > 0)
}
