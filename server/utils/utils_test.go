package utils

import (
	"fmt"
	"testing"

	tu "registro/utils/testutils"
)

func TestOffuscateur(t *testing.T) {
	offuscateur := Offuscateur[int64]{
		Prefix: "VI",
		M:      12,
		A:      9,
		B:      1,
	}
	for _, id := range []int64{4, 4568, 12, 2, 5, 7, 454, 78, 9899, 66656} {
		res, ok := offuscateur.Unmask(offuscateur.Mask(id))
		tu.Assert(t, ok)
		tu.Assert(t, res == id)
	}
	fmt.Println(offuscateur.Mask(1))
	fmt.Println(offuscateur.Mask(456))
	fmt.Println(offuscateur.Mask(15456))
}
