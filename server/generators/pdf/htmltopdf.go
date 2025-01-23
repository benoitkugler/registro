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

var (
	fc      text.FontConfiguration
	rootDir string
)

// Init setups font resources required for HTML to PDF conversion,
// and defines [root], which is the folder containing 'assets/'
// used in HTML templates.
func Init(fontCacheDir, root string) error {
	rootDir = root

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

// root is the folder containing 'assets/'
func htmlToPDF(html, root string) ([]byte, error) {
	var dst bytes.Buffer
	err := goweasyprint.HtmlToPdfOptions(&dst, utils.InputString(html), root, nil, "", nil, false, fc, 1, nil)
	return dst.Bytes(), err
}

func templateToPDF(t *template.Template, args any) ([]byte, error) {
	var html bytes.Buffer
	err := t.ExecuteTemplate(&html, "main.html", args)
	if err != nil {
		return nil, fmt.Errorf("generating html: %s", err)
	}
	return htmlToPDF(html.String(), rootDir)
}
