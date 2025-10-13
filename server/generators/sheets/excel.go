package sheets

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"registro/logic"
	cps "registro/sql/camps"

	"github.com/xuri/excelize/v2"
)

type Border uint8

const (
	Left Border = 1 << iota
	Right
	Top
	Bottom

	// BorderDouble is set to diplay a double border
	BorderDouble
)

type Alignement uint8

const (
	ACenter Alignement = iota + 1
	ALeft
	ARight
	AVertical
)

// returns an excel compatible string
func (a Alignement) String() string {
	switch a {
	case ACenter:
		return "center"
	case ALeft:
		return "left"
	case ARight:
		return "right"
	case AVertical:
		return "right"
	default:
		return ""
	}
}

type NumFormat uint8

const (
	_ NumFormat = iota
	Float
	Int
	Euros
	FrancsSuisse
	Percentage
)

// Style expose les attributs à appliquer sur une case.
// doit pouvoir être une clé de dictionnaire
type Style struct {
	Color          string // Hex string, such as #FF00AA
	Bold, Italic   bool
	Border         Border
	TextAlignement Alignement
	NumFormat      NumFormat
}

func newStyle(color string, bold, italic, withLeftBorder bool, format NumFormat) Style {
	b := Border(0)
	if withLeftBorder {
		b = Left | BorderDouble
	}
	s := Style{Color: color, Bold: bold, Italic: italic, Border: b, NumFormat: format}
	return s
}

func stringWidth(s string, bold bool) float64 {
	c := 1.1
	if bold {
		c = 1.2
	}
	return 2 + c*float64(len([]rune(s)))
}

func findColWidth(headers []string, liste [][]Cell, colIndex int) float64 {
	maxWidth := stringWidth(headers[colIndex], true)
	for _, row := range liste {
		if width := stringWidth(row[colIndex].Value, false); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

// Builder dirige la création d'un fichier excel.
// Les numéros de ligne et colonne commencent à 1.
type Builder struct {
	file *excelize.File

	styles map[[2]int]Style // (row, col) -> style

	err error
}

func NewBuilder() *Builder {
	return &Builder{file: excelize.NewFile(), styles: map[[2]int]Style{}}
}

// Finalize apply styles and returns the file contents
func (b *Builder) Finalize() (*bytes.Buffer, error) {
	if b.err != nil {
		return nil, b.err
	}

	err := b.applyStyles()
	if err != nil {
		return nil, err
	}
	return b.file.WriteToBuffer()
}

func (b *Builder) MergeCells(startRow, startCol, endRow, endCol int) {
	start, err := excelize.CoordinatesToCellName(startCol, startRow)
	if err != nil {
		b.err = err
		return
	}
	end, err := excelize.CoordinatesToCellName(endCol, endRow)
	if err != nil {
		b.err = err
		return
	}
	b.err = b.file.MergeCell("Sheet1", start, end)
}

// SetColumnWidth set the column width.
// See also [FitColsWidth]
func (b *Builder) SetColumnWidth(col int, width float64) {
	colLetter, err := excelize.ColumnNumberToName(col)
	if err != nil {
		b.err = err
		return
	}
	err = b.file.SetColWidth("Sheet1", colLetter, colLetter, width)
	if err != nil {
		b.err = err
		return
	}
}

func (b *Builder) SetCellF(row, col int, value float32) {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		b.err = err
		return
	}
	b.err = b.file.SetCellValue("Sheet1", cell, value)
}

func (b *Builder) SetCell(row, col int, value string) {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		b.err = err
		return
	}
	b.err = b.file.SetCellStr("Sheet1", cell, value)
}

// SetStyle enregistre le style pour la case donnée.
// Le style est effectivement appliqué par [Finalize].
func (b Builder) SetStyle(row, col int, style Style) { b.styles[[2]int{row, col}] = style }

// enregistre le style sur excelize
// ne devrait être appelé qu'une seule fois par style
func (b Builder) register(s Style) (int, error) {
	var excelS excelize.Style
	if s.Color != "" {
		excelS.Fill = excelize.Fill{Type: "pattern", Color: []string{s.Color}, Pattern: 1}
	}
	excelS.Font = &excelize.Font{Bold: s.Bold, Italic: s.Italic}
	if s.Border != 0 {
		var l []excelize.Border
		style := 1
		if s.Border&BorderDouble != 0 {
			style = 6
		}
		if s.Border&Left != 0 {
			l = append(l, excelize.Border{Type: "left", Color: "000000", Style: style})
		}
		if s.Border&Right != 0 {
			l = append(l, excelize.Border{Type: "right", Color: "000000", Style: style})
		}
		if s.Border&Top != 0 {
			l = append(l, excelize.Border{Type: "top", Color: "000000", Style: style})
		}
		if s.Border&Bottom != 0 {
			l = append(l, excelize.Border{Type: "bottom", Color: "000000", Style: style})
		}
		excelS.Border = l
	}

	if s.TextAlignement == AVertical {
		excelS.Alignment = &excelize.Alignment{
			Horizontal: ACenter.String(), Vertical: "center",
			TextRotation: 90, WrapText: true,
		}
	} else if s.TextAlignement != 0 {
		excelS.Alignment = &excelize.Alignment{Horizontal: s.TextAlignement.String()}
	}

	switch s.NumFormat {
	case Int:
		excelS.NumFmt = 1
	case Euros:
		excelS.NumFmt = 219
	case FrancsSuisse:
		excelS.NumFmt = 297
	case Percentage:
		excelS.NumFmt = 10
	}

	return b.file.NewStyle(&excelS)
}

func (b Builder) applyStyles() error {
	// unify styles
	m := map[Style]int{}
	for _, v := range b.styles {
		m[v] = 0
	}
	// register and convert to internal IDs
	for s := range m {
		id, err := b.register(s)
		if err != nil {
			return err
		}
		m[s] = id
	}
	// apply the correct style
	for cell, style := range b.styles {
		row, col := cell[0], cell[1]
		cellName, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			return err
		}
		if err = b.file.SetCellStyle("Sheet1", cellName, cellName, m[style]); err != nil {
			return err
		}
	}
	return nil
}

type oneTotal struct {
	label string
	value string
}

type Cell struct {
	Value  string
	ValueF float32 // used if [NumFormat] is not zero

	Color string // empty for no background color
	Bold  bool

	NumFormat NumFormat
}

func (b *Builder) drawItems(liste [][]Cell, startingRow int, showLineNumbers bool, colLeftBorder int) error {
	var colOffset int // pour les numéros de lignes
	if showLineNumbers {
		colOffset = 1
	}

	for row, data := range liste {
		currentRow := row + startingRow
		if showLineNumbers {
			b.SetCell(currentRow, 1, strconv.Itoa(row+1))
		}
		for col, cell := range data {
			currentCol := col + 1 + colOffset

			style := newStyle(cell.Color, cell.Bold, false, colLeftBorder == currentCol, cell.NumFormat)
			b.SetStyle(currentRow, currentCol, style)
			if cell.NumFormat != 0 {
				b.SetCellF(currentRow, currentCol, cell.ValueF)
			} else {
				b.SetCell(currentRow, currentCol, cell.Value)
			}
		}
	}
	return nil
}

// if `colLeftBorder == -1`, aucune ligne verticale supplémentaire n'est tracée
func renderListe(headers []string, liste [][]Cell, totals []oneTotal, showLineNumbers bool, colLeftBorder int) (*bytes.Buffer, error) {
	b := NewBuilder()
	var colOffset int // pour les numéros de lignes
	if showLineNumbers {
		colOffset = 1
	}

	// headers
	for col, field := range headers {
		b.SetCell(1, col+1+colOffset, field)

		b.SetStyle(1, col+1+colOffset, newStyle("", true, false, false, 0))
		colLetter, err := excelize.ColumnNumberToName(col + 1 + colOffset)
		if err != nil {
			return nil, err
		}
		colWidth := findColWidth(headers, liste, col)
		if err := b.file.SetColWidth("Sheet1", colLetter, colLetter, colWidth); err != nil {
			return nil, err
		}
	}

	// datas
	if err := b.drawItems(liste, 2, showLineNumbers, colLeftBorder); err != nil {
		return nil, err
	}

	// pour une ligne de totaux
	totalRow := len(liste) + 3
	for index, total := range totals {
		b.SetCell(totalRow, 2*index+1+colOffset, total.label)
		b.SetStyle(totalRow, 2*index+1+colOffset, newStyle("", false, true, false, 0))
		b.SetCell(totalRow, 2*index+2+colOffset, total.value)
		b.SetStyle(totalRow, 2*index+2+colOffset, newStyle("", true, false, false, 0))
	}

	return b.Finalize()
}

// CreateTable returns an Excel file for the basic data defined
// by [headers] and [liste]
func CreateTable(headers []string, liste [][]Cell) ([]byte, error) {
	f, err := renderListe(headers, liste, nil, false, -1)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

func CreateTableTotal(headers []string, liste [][]Cell, total string) ([]byte, error) {
	totals := []oneTotal{
		{"Total :", total},
	}
	f, err := renderListe(headers, liste, totals, true, -1)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

// SuiviFinancierCamp renvoie un tableau des participants avec l'état de leur facture
// Les champs suivants sont requis :
//   - FinancesPNomPrenom
//   - FinancesPPrixBase
//   - FinancesPPrixNet
//   - FinancesPTotalAides
//   - FinancesPEtatPaiement
func SuiviFinancierCamp(liste [][]Cell, totalDemande,
	totalAides string,
) ([]byte, error) {
	totals := []oneTotal{
		{"Total demandé:", totalDemande},
		{"Total aides:", totalAides},
	}
	headers := [...]string{
		"Participant",         // FinancesPNomPrenom
		"Prix de base (€)",    // FinancesPPrixBase
		"Montant attendu (€)", // FinancesPPrixNet
		"Dont aides (€)",      // FinancesPTotalAides
		"Etat du paiement",    // FinancesPEtatPaiement
	}
	f, err := renderListe(headers[:], liste, totals, false, -1)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

func formatBool(b bool) string {
	if b {
		return "Oui"
	}
	return "Non"
}

// formatTime returns a time following 22/02/2025 21:15:04 format
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d %02d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
}

// ListeParticipants renvoie un document Excel
func ListeParticipants(camp cps.Camp, inscrits []cps.ParticipantPersonne, dossiers logic.Dossiers, groupes map[cps.IdParticipant]cps.Groupe,
	showNationnaliteSuisse bool,
) ([]byte, error) {
	headersParticipant := [...]string{
		"Inscription",
		"Nom",
		"Prénom",
		"Sexe",
		"Date de naissance",
		"Age (début de camp)",
		"Mail du participant",
		"Groupe",
		"Navette",
		"Commentaire",
		"", // hidden if showNationnaliteSuisse is false
	}
	if showNationnaliteSuisse {
		headersParticipant[10] = "Nationalité suisse"
	}

	headersResponsable := [...]string{
		"Responsable",
		"Mail",
		"Tel.",
		"Adresse",
		"Code postal",
		"Ville",
		"Pays",
	}

	headers := append(headersParticipant[:], headersResponsable[:]...)
	colLine := len(headersParticipant) + 1

	rows := make([][]Cell, len(inscrits))
	for i, inscrit := range inscrits {
		dossier := dossiers.For(inscrit.Participant.IdDossier)
		responsable := dossier.Responsable()
		groupe := groupes[inscrit.Participant.Id]
		nationalite := ""
		if showNationnaliteSuisse {
			nationalite = formatBool(inscrit.Personne.Nationnalite.IsSuisse)
		}
		var row [len(headersParticipant) + len(headersResponsable)]Cell = [...]Cell{
			// inscrit
			{Value: formatTime(dossier.Dossier.MomentInscription)},                               // Inscription
			{Value: inscrit.Personne.FNom()},                                                     // Nom
			{Value: inscrit.Personne.FPrenom()},                                                  // Prénom
			{Value: inscrit.Personne.Sexe.String()},                                              // Sexe
			{Value: inscrit.Personne.DateNaissance.String()},                                     // Date de naissance
			{ValueF: float32(camp.AgeDebutCamp(inscrit.Personne.DateNaissance)), NumFormat: Int}, // Age (début de camp)
			{Value: inscrit.Personne.Mail},                                                       // Mail du participant
			{Value: groupe.Nom, Color: groupe.Couleur},                                           // Groupe
			{Value: inscrit.Participant.Navette.String()},                                        // Navette
			{Value: inscrit.Participant.Commentaire},                                             // Commentaire
			{Value: nationalite},                                                                 // Suisse ?
			// responsable
			{Value: responsable.NOMPrenom()},   // Responsable
			{Value: responsable.Mail},          // Mail
			{Value: responsable.Tels.String()}, // Tel.
			{Value: responsable.Adresse},       // Adresse
			{Value: responsable.CodePostal},    // Code postal
			{Value: responsable.Ville},         // Ville
			{Value: string(responsable.Pays)},  // Pays
		}
		rows[i] = row[:]
	}

	f, err := renderListe(headers, rows, nil, false, colLine)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}
