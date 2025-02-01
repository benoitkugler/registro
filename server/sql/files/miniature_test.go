package files

import (
	"bytes"
	"os"
	"testing"
	"time"

	tu "registro/utils/testutils"
)

func TestMiniature(t *testing.T) {
	tu.AssertNoErr(t, Init())

	f, err := os.Open("test/img1.png")
	tu.AssertNoErr(t, err)
	min, err := ComputeMiniature(".png", f)
	tu.AssertNoErr(t, err)
	tu.Write(t, "min1.png", min)

	f, err = os.Open("test/img2.JPG")
	tu.AssertNoErr(t, err)
	min, err = ComputeMiniature(".JPG", f)
	tu.AssertNoErr(t, err)
	tu.Write(t, "min2.png", min)

	bs, err := os.ReadFile("test/doc3.pdf")
	tu.AssertNoErr(t, err)
	min, err = ComputeMiniature(".pdf", bytes.NewReader(bs))
	tu.AssertNoErr(t, err)
	tu.Write(t, "min3.png", min)

	minAlt, err := ComputeMiniaturePDF(bs)
	tu.AssertNoErr(t, err)
	tu.Assert(t, bytes.Equal(min, minAlt))

	_, err = ComputeMiniature(".xxx", nil)
	tu.AssertErr(t, err)
	_, err = ComputeMiniature(".png", &bytes.Reader{})
	tu.AssertErr(t, err)
}

func TestFilepath(t *testing.T) {
	tu.Assert(t, IdFile(4).filepath("root", false) == "root/file_4")
	tu.Assert(t, IdFile(4).filepath("root", true) == "root/file_4_min")
}

func TestNewFile(t *testing.T) {
	tu.Assert(t, NewFile(nil, "").Taille == 0)
	tu.Assert(t, NewFile(nil, "").DateHeureModif.Day() == time.Now().Day())
}
