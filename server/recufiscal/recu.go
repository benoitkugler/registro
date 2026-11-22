package recufiscal

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

var recuTemplate *model.Document

// Init loads the PDF template required to emit "recu fiscal"
func Init(templatePath string) error {
	doc, _, err := reader.ParsePDFFile(templatePath, reader.Options{})
	if err != nil {
		return fmt.Errorf("loading PDF template: %s", err)
	}
	if L := len(doc.Catalog.AcroForm.Flatten()); L != 36 {
		return fmt.Errorf("loading PDF template: corrupted model (unexpected number of form fields)")
	}
	field := doc.Catalog.AcroForm.Flatten()["numeroRecu"].Field // pointer
	field.FT = model.FormFieldText{}                            // remove max length limitation

	recuTemplate = &doc
	return nil
}

type recuFiscal struct {
	// Montant total
	Montant ds.MontantTaux

	// Dans le cas de plusieurs dons, le mode est un des modes utilisés
	Mode ds.ModePaiement
	// Dans le cas de plusieurs dons, la date la plus récente est utilisée
	Date time.Time
}

func loadAndSelect(db dons.DB, year int) (map[pr.IdPersonne]recuFiscal, pr.Personnes, error) {
	dons, err := dons.SelectAllDons(db)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}

	// selectForRecu aggrège les dons, ignorant les dons collectifs.
	// TODO: support for custom taux, not needed for now since Repere does not use this API
	taux := ds.Taux{Euros: 1000}
	selected := map[pr.IdPersonne]recuFiscal{}
	for _, don := range dons {
		// les dons collectifs ne sont pas concernés par les reçus fiscaux
		if !don.IdPersonne.Valid {
			continue
		}
		if don.Date.Time().Year() != year { // restrict to selected year
			continue
		}

		rf, ok := selected[don.IdPersonne.Id]
		if !ok {
			rf.Montant = taux.Zero()
		}

		rf.Montant.Add(don.Montant)
		rf.Mode = don.ModePaiement
		if d := don.Date.Time(); rf.Date.Before(d) {
			rf.Date = d
		}
		selected[don.IdPersonne.Id] = rf
	}

	idDonateurs := utils.MapKeys(selected)

	personnes, err := pr.SelectPersonnes(db, idDonateurs...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	return selected, personnes, nil
}

func generate(recus map[pr.IdPersonne]recuFiscal, personnes pr.Personnes) ([]byte, error) {
	var archiveB bytes.Buffer
	archive := utils.NewZip(&archiveB)

	etiquettes := make([]pr.Identite, 0, len(recus))

	for idDonateur, r := range recus {
		donateur := personnes[idDonateur]
		donPDF, err := fillPdf(r, donateur)
		if err != nil {
			return nil, err
		}
		archive.AddFile(fmt.Sprintf("recu_%s_%d.pdf", donateur.NOMPrenom(), idDonateur), bytes.NewReader(donPDF))
		etiquettes = append(etiquettes, donateur.Identite)
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

// Generate rassemble les dons (personnels, pour l'année donnée) et
// renvoie une archive .ZIP contenant :
//   - les étiquettes,
//   - les reçus fiscaux
//   - les adresses mails dans un fichier .CSV
func Generate(db dons.DB, year int) ([]byte, error) {
	if recuTemplate == nil { // better error than panic
		return nil, errors.New("internal error: recu template was not initialized")
	}

	selected, personnes, err := loadAndSelect(db, year)
	if err != nil {
		return nil, err
	}

	return generate(selected, personnes)
}

// indique l'identifiant des champs présents dans le modèle.
// A synchroniser avec modele_2026_form.pdf
type champPdf struct {
	id     string
	valeur formfill.FDFValue
}

func champsAsso(asso association) []champPdf {
	categorie := formfill.FDFName("Off")
	if asso.ReconnueUtilitePublique {
		categorie = formfill.FDFName("Oui")
	}

	return []champPdf{
		{id: "nomOrganisme", valeur: formfill.FDFText(asso.Nom)},
		{id: "numeroSIREN", valeur: formfill.FDFText(asso.NumeroSIREN)},
		{id: "adresseOrganisme", valeur: formfill.FDFText(asso.Adresse)},
		{id: "codePostalOrganisme", valeur: formfill.FDFText(asso.CodePostal)},
		{id: "villeOrganisme", valeur: formfill.FDFText(asso.Ville)},
		{id: "paysOrganisme", valeur: formfill.FDFText(asso.Pays)},
		{id: "objectifLigne1", valeur: formfill.FDFText(asso.ObjectifLigne1)},
		{id: "objectifLigne2", valeur: formfill.FDFText(asso.ObjectifLigne2)},

		// Catégorie de l'association
		{id: "associationReconnueUtilitePublique", valeur: categorie},

		{id: "jourDecret", valeur: formfill.FDFText(asso.Decret.Jour)},
		{id: "moisDecret", valeur: formfill.FDFText(asso.Decret.Mois)},
		{id: "anneeDecret", valeur: formfill.FDFText(asso.Decret.Annee)},
		{id: "jourJournalOfficiel", valeur: formfill.FDFText(asso.JournalOfficiel.Jour)},
		{id: "moisJournalOfficiel", valeur: formfill.FDFText(asso.JournalOfficiel.Mois)},
		{id: "anneeJournalOfficiel", valeur: formfill.FDFText(asso.JournalOfficiel.Annee)},
	}
}

func champsDonateur(donateur pr.Personne) []champPdf {
	return []champPdf{
		{id: "nomDonateur", valeur: formfill.FDFText(donateur.FNom())},
		{id: "prenomDonateur", valeur: formfill.FDFText(donateur.FPrenom())},
		{id: "adresseDonateur", valeur: formfill.FDFText(donateur.Adresse)},
		{id: "codePostalDonateur", valeur: formfill.FDFText(donateur.CodePostal)},
		{id: "villeDonateur", valeur: formfill.FDFText(strings.ToUpper(donateur.Ville))},
		{id: "paysDonateur", valeur: formfill.FDFText(string(donateur.Pays))},
	}
}

// `don` représente une aggrégation de plusieurs dons.
func champsDon(don recuFiscal) []champPdf {
	date := don.Date
	euros := don.Montant.Convert(ds.Euros)
	montantLettre := num2words.EurosToWords(euros.Cent)
	modeDon := map[ds.ModePaiement]string{
		ds.Especes:   "modeVersementEspeces",
		ds.Cheque:    "modeVersementCheque",
		ds.Virement:  "modeVersementVirement",
		ds.Helloasso: "modeVersementVirement",
		ds.EnLigne:   "modeVersementVirement",
		ds.Ancv:      "modeVersementCheque",
	}
	return []champPdf{
		{id: "montantChiffres", valeur: formfill.FDFText(euros.String())},
		{id: "montantLettres", valeur: formfill.FDFText(montantLettre)},
		{id: "jourVersement", valeur: formfill.FDFText(strconv.Itoa(date.Day()))},
		{id: "moisVersement", valeur: formfill.FDFText(strconv.Itoa(int(date.Month())))},
		{id: "anneeVersement", valeur: formfill.FDFText(strconv.Itoa(date.Year()))},
		{id: modeDon[don.Mode], valeur: formfill.FDFName("Oui")},
	}
}

var champsTypeDon = []champPdf{
	{id: "article200CGI", valeur: formfill.FDFName("Oui")},
	{id: "natureDonNumeraire", valeur: formfill.FDFName("Oui")},
	{id: "formeDonManuel", valeur: formfill.FDFName("Oui")},
}

func champsDateEdition() []champPdf {
	today := time.Now()
	return []champPdf{
		{id: "jourEdition", valeur: formfill.FDFText(strconv.Itoa(today.Day()))},
		{id: "moisEdition", valeur: formfill.FDFText(strconv.Itoa(int(today.Month())))},
		{id: "anneeEdition", valeur: formfill.FDFText(strconv.Itoa(today.Year()))},
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
		{id: "numeroRecu", valeur: formfill.FDFText(numero(donateur.Id))},
	}
	fields = append(fields, champsAsso(assoACVE)...)
	fields = append(fields, champsDonateur(donateur)...)
	fields = append(fields, champsDon(recu)...)
	fields = append(fields, champsTypeDon...)
	fields = append(fields, champsDateEdition()...)

	var pdfFields []formfill.FDFField
	for _, field := range fields {
		pdfFields = append(pdfFields, formfill.FDFField{T: field.id, Values: formfill.Values{V: field.valeur}})
	}

	doc := recuTemplate.Clone()
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
