package recufiscal

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"time"

	"registro/generators/sheets"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/benoitkugler/num2words"
	"github.com/benoitkugler/pdf/formfill"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader"
)

//go:embed templateRecuAcve.pdf
var templateRecuAcve []byte

var docAcve model.Document

// load the PDF model
func init() {
	doc, _, err := reader.ParsePDFReader(bytes.NewReader(templateRecuAcve), reader.Options{})
	if err != nil {
		panic(err)
	}
	if L := len(doc.Catalog.AcroForm.Flatten()); L != 65 {
		panic("corrupted model (unexpected number of form fields)")
	}
	field := doc.Catalog.AcroForm.Flatten()["z1"].Field
	field.FT = model.FormFieldText{} // remove max length limitation

	docAcve = doc
}

type recuFiscal struct {
	// Montant total
	Montant ds.MontantTaux

	// Dans le cas de plusieurs dons, le mode est un des modes utilisés
	Mode ds.ModePaiement
	// Dans le cas de plusieurs dons, la date la plus récente est utilisée
	Date time.Time
}

// selectForRecu aggrège les dons, ignorant les dons collectifs.
func selectForRecu(taux ds.Taux, dons dons.Dons, year int) map[pr.IdPersonne]recuFiscal {
	out := map[pr.IdPersonne]recuFiscal{}
	for _, don := range dons {
		// les dons collectifs ne sont pas concernés par les reçus fiscaux
		if !don.IdPersonne.Valid {
			continue
		}
		if don.Date.Time().Year() != year { // restrict to selected year
			continue
		}

		rf, ok := out[don.IdPersonne.Id]
		if !ok {
			rf.Montant = taux.Zero()
		}

		rf.Montant.Add(don.Montant)
		rf.Mode = don.ModePaiement
		if d := don.Date.Time(); rf.Date.Before(d) {
			rf.Date = d
		}
		out[don.IdPersonne.Id] = rf
	}

	return out
}

// Generate rassemble les dons (personnels, pour l'année donnée) et
// renvoie une archive .zip contenant :
//   - les étiquettes,
//   - les reçus fiscaux
//   - les adresses mails dans un CSV
func Generate(db dons.DB, year int) ([]byte, error) {
	dons, err := dons.SelectAllDons(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// TODO: support for custom taux
	selected := selectForRecu(ds.Taux{Euros: 1000}, dons, year)
	idDonateurs := utils.MapKeys(selected)

	personnes, err := pr.SelectPersonnes(db, idDonateurs...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	var archiveB bytes.Buffer
	archive := utils.NewZip(&archiveB)

	etiquettes := make([]pr.Etatcivil, 0, len(selected))

	for idDonateur, r := range selected {
		donateur := personnes[idDonateur]
		donPDF, err := fillPdf(r, donateur)
		if err != nil {
			return nil, err
		}
		archive.AddFile(fmt.Sprintf("recu_%s_%d.pdf", donateur.NOMPrenom(), idDonateur), bytes.NewReader(donPDF))
		etiquettes = append(etiquettes, donateur.Etatcivil)
	}

	// sort etiquettes by name
	sort.Slice(etiquettes, func(i, j int) bool {
		return etiquettes[i].NOMPrenom() < etiquettes[j].NOMPrenom()
	})
	etiquettesPDF, err := genereEtiquettes(etiquettes)
	if err != nil {
		return nil, err
	}
	archive.AddFile("etiquettes.pdf", bytes.NewReader(etiquettesPDF))

	// mails
	liste := make([][]string, len(etiquettes)+1)
	liste[0] = []string{"Nom", "Prénom", "Mail"}
	for i, e := range etiquettes {
		liste[i+1] = []string{e.FNom(), e.FPrenom(), e.Mail}
	}
	mailsCSV, err := sheets.CreateCsv(liste)
	if err != nil {
		return nil, err
	}
	archive.AddFile("mails.csv", bytes.NewReader(mailsCSV))

	err = archive.Close()
	if err != nil {
		return nil, err
	}

	return archiveB.Bytes(), nil
}

// indique l'identifiant des champs présents dans le modèle.
// A synchroniser avec ModeleRecuFiscalEditable.pdf
type champPdf struct {
	id     string
	valeur formfill.FDFValue
}

var champsACVE = []champPdf{
	// ACVE
	{id: "z2" /* nom */, valeur: formfill.FDFText("ACVE")},
	{id: "z4" /* adresse */, valeur: formfill.FDFText("La Maison du Rocher")},
	{id: "z5" /* codePostal */, valeur: formfill.FDFText("26160")},
	{id: "z5b" /* ville */, valeur: formfill.FDFText("CHAMALOC")},
	{id: "z6" /* objectifL1 */, valeur: formfill.FDFText("Créer et gérer des séjours pour enfants, adolescents et adultes.")},
	{id: "z7" /* objectifL2 */, valeur: formfill.FDFText("Faire connaître, à travers des animations adaptées à l’âge des participants, les valeurs chrétiennes.")},

	// CATEGORIE ASSO
	{id: "z9", valeur: formfill.FDFName("Oui")},

	{id: "d3" /* anneeDecret */, valeur: formfill.FDFText("1957")},
	{id: "d3b" /* anneeJournal */, valeur: formfill.FDFText("1957")},
	{id: "d1" /* jourDecret */, valeur: formfill.FDFText("5")},
	{id: "d1b" /* jourJournal */, valeur: formfill.FDFText("29")},
	{id: "d2" /* moisDecret */, valeur: formfill.FDFText("1")},
	{id: "d2b" /* moisJournal */, valeur: formfill.FDFText("1")},
}

func champsDonateur(donateur pr.Personne) []champPdf {
	return []champPdf{
		{id: "z29" /* nom */, valeur: formfill.FDFText(donateur.Nom)},
		{id: "z30" /* prenom */, valeur: formfill.FDFText(donateur.Prenom)},
		{id: "z31" /* adresse */, valeur: formfill.FDFText(donateur.Adresse)},
		{id: "z32" /* codePostal */, valeur: formfill.FDFText(donateur.CodePostal)},
		{id: "z33" /* ville */, valeur: formfill.FDFText(donateur.Ville)},
	}
}

// `don` représente une aggrégation de plusieurs dons.
func champsDon(don recuFiscal) []champPdf {
	date := don.Date
	euros := don.Montant.Convert(ds.Euros)
	montantLettre := num2words.EurosToWords(euros.Cent)
	modeDon := map[ds.ModePaiement]string{
		ds.Especes:   "z49",
		ds.Cheque:    "z50",
		ds.Virement:  "z51",
		ds.Helloasso: "z51",
		ds.EnLigne:   "z51",
		ds.Ancv:      "z50",
	}
	return []champPdf{
		{id: "z34" /* montantChiffre */, valeur: formfill.FDFText(euros.String())},
		{id: "z35" /* montantLettre */, valeur: formfill.FDFText(montantLettre)},
		{id: "z36" /* jourVersement */, valeur: formfill.FDFText(strconv.Itoa(date.Day()))},
		{id: "z37" /* moisVersement */, valeur: formfill.FDFText(strconv.Itoa(int(date.Month())))},
		{id: "z38" /* anneeVersement */, valeur: formfill.FDFText(strconv.Itoa(date.Year()))},
		{id: modeDon[don.Mode], valeur: formfill.FDFName("Oui")},
	}
}

var champsTypeDon = []champPdf{
	{id: "z39" /* articleLoi[200] */, valeur: formfill.FDFName("Oui")},
	{id: "z46" /* natureDon[manuel] */, valeur: formfill.FDFName("Oui")},
	{id: "z44" /* formeDon[numeraire] */, valeur: formfill.FDFName("Oui")},
}

func champsDateEdition() []champPdf {
	today := time.Now()
	return []champPdf{
		{id: "z52" /* jour */, valeur: formfill.FDFText(strconv.Itoa(today.Day()))},
		{id: "z53" /* mois */, valeur: formfill.FDFText(strconv.Itoa(int(today.Month())))},
		{id: "z54" /* annee */, valeur: formfill.FDFText(strconv.Itoa(today.Year()))},
	}
}

// renvoie le numéro de reçu (unique par donateur / année)
func numero(idPersonne pr.IdPersonne) string {
	n := time.Now()
	return fmt.Sprintf("%04d%02d", idPersonne, n.Year()%100)
}

// fillPdf fills the .PDF template with the given fields and
// returns the PDF bytes.
func fillPdf(recu recuFiscal, donateur pr.Personne) ([]byte, error) {
	fields := []champPdf{
		{id: "z1" /* numero */, valeur: formfill.FDFText(numero(donateur.Id))},
	}
	fields = append(fields, champsACVE...)
	fields = append(fields, champsDonateur(donateur)...)
	fields = append(fields, champsDon(recu)...)
	fields = append(fields, champsTypeDon...)
	fields = append(fields, champsDateEdition()...)

	var pdfFields []formfill.FDFField
	for _, field := range fields {
		pdfFields = append(pdfFields, formfill.FDFField{T: field.id, Values: formfill.Values{V: field.valeur}})
	}

	doc := docAcve.Clone()
	err := formfill.FillForm(&doc, formfill.FDFDict{Fields: pdfFields}, true)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = doc.Write(&buf, nil)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
