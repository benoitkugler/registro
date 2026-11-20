package compta

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	tu "registro/utils/testutils"

	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader"
)

func TestMergeImagesToPDF(t *testing.T) {
	first := pngTestImage(20, 30)
	second := pngTestImage(50, 40)

	pdf, err := mergeImagesToPDF([]justificatifUpload{
		{content: first, filename: "test.png"},
		{content: second, filename: "test.png"},
	})
	tu.AssertNoErr(t, err)

	tu.Write(t, "test_image_merged.pdf", pdf)

	doc, _, err := reader.ParsePDFReader(bytes.NewReader(pdf), reader.Options{})
	tu.AssertNoErr(t, err)
	pages := doc.Catalog.Pages.Flatten()
	tu.Assert(t, len(pages) == 2)

	for i, expected := range [][2]int{{20, 30}, {50, 40}} {
		var found *model.XObjectImage
		for _, xObject := range pages[i].Resources.XObject {
			if img, ok := xObject.(*model.XObjectImage); ok {
				found = img
				break
			}
		}
		tu.Assert(t, found != nil)
		tu.Assert(t, found.Width == expected[0] && found.Height == expected[1])
	}
}

func TestImageUploadMIMERejectsInvalidOrMismatchedImages(t *testing.T) {
	_, err := imageUploadMIME(justificatifUpload{filename: "document.png", content: []byte("not an image")})
	tu.AssertErr(t, err)

	_, err = imageUploadMIME(justificatifUpload{filename: "document.jpg", content: pngTestImage(2, 3)})
	tu.AssertErr(t, err)

	_, err = imageUploadMIME(justificatifUpload{filename: "document.pdf"})
	tu.AssertErr(t, err)
}

func pngTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: uint8(y), A: 255})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		panic(err)
	}
	return out.Bytes()
}
