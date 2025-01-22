package pdfcreator

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	goweasyprint "github.com/benoitkugler/go-weasyprint"
	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

var fc text.FontConfiguration

// Init setup font resources required for HTML to PDF conversion.
func Init(fontCacheDir string) error {
	fontcache := filepath.Join(fontCacheDir, "font.cache")

	fs, err := fontconfig.LoadFontsetFile(fontcache)
	if err != nil {
		_, err := fontconfig.ScanAndCache(fontcache)
		if err != nil {
			return err
		}
		fs, err = fontconfig.LoadFontsetFile(fontcache)
		if err != nil {
			return err
		}
	}

	fc = text.NewFontConfigurationPango(fcfonts.NewFontMap(fontconfig.Standard, fs))
	return nil
}

func htmlToPDF(html string) ([]byte, error) {
	var dst bytes.Buffer
	// TODO : proper asset filepath
	err := goweasyprint.HtmlToPdfOptions(&dst, utils.InputString(html), "../..", nil, "", nil, true, fc, 1, nil)
	return dst.Bytes(), err
}

func templateToPDF(t *template.Template, args any) ([]byte, error) {
	var html bytes.Buffer
	err := t.ExecuteTemplate(&html, "main.html", args)
	if err != nil {
		return nil, fmt.Errorf("generating html: %s", err)
	}
	return htmlToPDF(html.String())
}
