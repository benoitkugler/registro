package compta

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/model"
)

const maxJustificatifSize = 10 * 1024 * 1024

type justificatifUpload struct {
	filename string
	content  []byte
}

func imageUploadMIME(upload justificatifUpload) (string, error) {
	extension := strings.ToLower(filepath.Ext(upload.filename))
	expectedFormat := ""
	switch extension {
	case ".png":
		expectedFormat = "png"
	case ".jpg", ".jpeg":
		expectedFormat = "jpeg"
	default:
		return "", fmt.Errorf("unsupported image format")
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(upload.content))
	if err != nil {
		return "", fmt.Errorf("invalid image %q: %w", upload.filename, err)
	}
	if format != expectedFormat {
		return "", fmt.Errorf("image extension does not match content for %q", upload.filename)
	}
	if format == "png" {
		return "image/png", nil
	}
	return "image/jpeg", nil
}

// assume images is not empty
func mergeImagesToPDF(uploads []justificatifUpload) ([]byte, error) {
	var doc model.Document
	for _, upload := range uploads {
		mimeType, err := imageUploadMIME(upload)
		if err != nil {
			return nil, err
		}
		img, _, err := contentstream.ParseImage(bytes.NewReader(upload.content), mimeType)
		if err != nil {
			return nil, fmt.Errorf("parsing uploaded image: %w", err)
		}

		pageWidth, pageHeight := model.Fl(595.28), model.Fl(841.89)
		if img.Width > img.Height {
			pageWidth, pageHeight = pageHeight, pageWidth
		}
		const margin = model.Fl(18)
		widthScale := (pageWidth - 2*margin) / model.Fl(img.Width)
		heightScale := (pageHeight - 2*margin) / model.Fl(img.Height)
		scale := widthScale
		if heightScale < scale {
			scale = heightScale
		}
		imageWidth := model.Fl(img.Width) * scale
		imageHeight := model.Fl(img.Height) * scale
		x := (pageWidth - imageWidth) / 2
		y := (pageHeight - imageHeight) / 2

		bounds := model.Rectangle{Urx: pageWidth, Ury: pageHeight}
		stream := contentstream.NewGraphicStream(bounds)
		stream.AddXObjectDims(img, x, y, imageWidth, imageHeight)
		page := new(model.PageObject)
		stream.ApplyToPageObject(page, true)
		doc.Catalog.Pages.Kids = append(doc.Catalog.Pages.Kids, page)
	}

	var out bytes.Buffer
	if err := doc.Write(&out, nil); err != nil {
		return nil, fmt.Errorf("writing merged PDF: %w", err)
	}
	return out.Bytes(), nil
}
