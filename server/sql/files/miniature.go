package files

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

const (
	miniatureWidth  = 200 // in pixels
	miniatureHeight = 300 // in pixels
)

// Init checks that Ghostscript is properly installed
func Init() error {
	err := exec.Command("gs", "--help").Run()
	if err != nil {
		return fmt.Errorf("command Ghostscript not found: %s", err)
	}
	return nil
}

// computeMiniature réduit le document entrant à une image png.
// Les formats supportés sont .pdf, .png, .jpg, .jpeg
//
// Le format .pdf requiert l'utilisation de Ghostscript.
func computeMiniature(extension string, doc io.Reader) ([]byte, error) {
	var miniature bytes.Buffer
	ext := strings.ToLower(filepath.Ext(extension))
	switch ext {
	case ".pdf":
		cmd := exec.Command("gs", "-sDEVICE=png16m", "-dFirstPage=1", "-dLastPage=1", "-sstdout=%stderr", "-sOutputFile=-",
			"-dBATCH", "-q", "-dTextAlphaBits=4", "-dGraphicsAlphaBits=4", "-dPDFFitPage", fmt.Sprintf("-g%dx%d", miniatureWidth, miniatureHeight), "-")
		cmd.Stdin = doc
		cmd.Stdout = &miniature
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("creating PDF miniature : %s", err)
		}
	case ".png", ".jpg", ".jpeg":
		srcImage, _, err := image.Decode(doc)
		if err != nil {
			return nil, fmt.Errorf("creating image miniature : %s", err)
		}
		if srcImage.Bounds().Dx() >= srcImage.Bounds().Dy() {
			srcImage = imaging.Rotate270(srcImage)
		}
		dstImageFit := imaging.Thumbnail(srcImage, miniatureWidth, miniatureHeight, imaging.Lanczos)
		if err = imaging.Encode(&miniature, dstImageFit, imaging.PNG); err != nil {
			return nil, fmt.Errorf("creating image miniature : %s", err)
		}
	default:
		return nil, fmt.Errorf("creating miniature : unsupported format %s", ext)
	}

	return miniature.Bytes(), nil
}

// ComputeMiniaturePDF est un raccourci pour les fichier PDF (voir ComputeMiniature)
func ComputeMiniaturePDF(pdfBytes []byte) (miniature []byte, err error) {
	return computeMiniature(".pdf", bytes.NewReader(pdfBytes))
}
