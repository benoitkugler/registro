package sheets

import (
	"testing"

	tu "registro/utils/testutils"
)

func TestCreateCsv(t *testing.T) {
	content, err := CreateCsv([][]string{
		{"a", "b", "c"},
		{"é", "32.5", "c"},
	})
	tu.AssertNoErr(t, err)

	tu.Write(t, "registro_test.csv", content)
}
