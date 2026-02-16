package recufiscal

import (
	"bytes"
	_ "embed"
	"html/template"
	"math"
	"strings"

	"registro/generators/pdfcreator"
	pr "registro/sql/personnes"
)

//go:embed etiquettes.html
var etiquetteTemplate string

var etiquettesT = template.Must(template.New("").Parse(etiquetteTemplate))

// genereEtiquettes	renvoie un pdf
func genereEtiquettes(personnes []pr.Identite) ([]byte, error) {
	type adresse struct {
		Personne   string
		Adresse    string
		CodePostal string
		Ville      string
	}
	rowsCount := int(math.Ceil(float64(len(personnes)) / 3))
	data := make([][]adresse, rowsCount) // rows and cells
	for i, p := range personnes {
		rowNumber := i / 3
		data[rowNumber] = append(data[rowNumber], adresse{
			Personne:   p.NOMPrenom(),
			Adresse:    p.Adresse,
			CodePostal: p.CodePostal,
			Ville:      strings.ToUpper(p.Ville),
		})
	}
	var out bytes.Buffer
	err := etiquettesT.Execute(&out, data)
	if err != nil {
		return nil, err
	}
	return pdfcreator.HTMLToPDF(out.String())
}
