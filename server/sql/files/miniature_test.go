package files

import (
	"bytes"
	"os"
	"testing"

	tu "registro/utils/testutils"
)

func TestMiniature(t *testing.T) {
	tu.AssertNoErr(t, Init())

	f, err := os.Open("test/img1.png")
	tu.AssertNoErr(t, err)
	min, err := computeMiniature(".png", f)
	tu.AssertNoErr(t, err)
	tu.Write(t, "min1.png", min)

	f, err = os.Open("test/img2.JPG")
	tu.AssertNoErr(t, err)
	min, err = computeMiniature(".JPG", f)
	tu.AssertNoErr(t, err)
	tu.Write(t, "min2.png", min)

	bs, err := os.ReadFile("test/doc3.pdf")
	tu.AssertNoErr(t, err)
	min, err = computeMiniature(".pdf", bytes.NewReader(bs))
	tu.AssertNoErr(t, err)
	tu.Write(t, "min3.png", min)

	minAlt, err := ComputeMiniaturePDF(bs)
	tu.AssertNoErr(t, err)
	tu.Assert(t, bytes.Equal(min, minAlt))

	_, err = computeMiniature(".xxx", nil)
	tu.AssertErr(t, err)
	_, err = computeMiniature(".png", &bytes.Reader{})
	tu.AssertErr(t, err)
}
