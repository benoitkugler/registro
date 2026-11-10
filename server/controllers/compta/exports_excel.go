package compta

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"registro/generators/sheets"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	sh "registro/sql/shared"
)

// excel support

type ecritures struct {
	taux        ds.Taux
	camp        string
	details     string
	ventilation ventilations
}

func (ld *comptaLoader) ecritures() ecritures {
	return ecritures{
		taux:        ld.taux,
		camp:        string(ld.camp.Label()),
		details:     fmt.Sprintf("%s %s - %s", ld.directeur.FPrenom(), ld.directeur.FNom(), accountLabels[ld.compta.Account]),
		ventilation: ld.aggregateDepenses(ld.resolveFraismdr()),
	}
}

type ecritureComptable struct {
	NumeroCompte string
	Intitule     string
	IsCredit     bool
	Montant      float32 // in euros
}

// only use the dedicated camp account
// toBalance is computed so that debit == credit
func (ec ecritures) compile() ([]ecritureComptable, float32) {
	sum := ec.ventilation.sum()
	debitTotal, receipes := sum.accCampDepenses.AggTo(ec.taux, ds.Euros), sum.accAssoReceipes.AggTo(ec.taux, ds.Euros)
	toBalance := debitTotal - receipes

	expList := make([]ecritureComptable, 0, len(ec.ventilation))
	for k, dep := range ec.ventilation {
		montant := dep.accCampDepenses.Agg(ec.taux)
		if montant.Cent == 0 {
			continue
		}
		item := ventilationLabels[k]
		expList = append(expList, ecritureComptable{item.codeACVE, item.label, false, float32(montant.Cent) / 100})
	}
	slices.SortFunc(expList, func(a, b ecritureComptable) int { return strings.Compare(a.Intitule, b.Intitule) })

	// add the receipes
	recList := make([]ecritureComptable, 0, len(ec.ventilation))
	for k, dep := range ec.ventilation {
		montant := dep.accCampReceipes.Agg(ec.taux)
		if montant.Cent == 0 {
			continue
		}
		item := ventilationLabels[k]
		recList = append(recList, ecritureComptable{"REC-" + item.codeACVE, "Recette - " + item.label, true, float32(montant.Cent) / 100})
	}
	slices.SortFunc(recList, func(a, b ecritureComptable) int { return strings.Compare(a.Intitule, b.Intitule) })

	out := slices.Concat(expList, recList, []ecritureComptable{
		{"5125", ec.camp, true, float32(toBalance) / 100},
	})

	return out, float32(debitTotal) / 100
}

func generateExcelEcritures(ec ecritures) (*bytes.Buffer, error) {
	b := sheets.NewBuilder()

	// header
	const startRow = 4
	for i, label := range [4]string{"Numéro", "Intitulé", "Débit", "Crédit"} {
		b.SetCell(startRow, i+1, label)
		b.SetStyle(startRow, i+1, sheets.Style{Color: "#c5e0b4", Bold: true, Border: sheets.Top | sheets.Bottom | sheets.BorderDouble, TextAlignement: sheets.ACenter})
	}

	// rows
	rows, tDebit := ec.compile()
	for j, item := range rows {
		row := startRow + 2 + j
		b.SetCell(row, 1, item.NumeroCompte)
		b.SetCell(row, 2, item.Intitule)
		style := sheets.Style{TextAlignement: sheets.ARight, NumFormat: sheets.Euros}
		if item.IsCredit {
			b.SetCellF(row, 4, item.Montant)
			b.SetStyle(row, 4, style)
		} else {
			b.SetCellF(row, 3, item.Montant)
			b.SetStyle(row, 3, style)
		}
	}

	// total
	totalRow := startRow + 2 + len(rows)
	b.SetCell(totalRow, 2, "Total")
	b.SetCellF(totalRow, 3, tDebit)
	b.SetCellF(totalRow, 4, tDebit)
	b.SetStyle(totalRow, 2, sheets.Style{Bold: true, TextAlignement: sheets.ARight})
	b.SetStyle(totalRow, 3, sheets.Style{Bold: true, TextAlignement: sheets.ARight})
	b.SetStyle(totalRow, 4, sheets.Style{Bold: true, TextAlignement: sheets.ARight})

	b.FitColsWidth()

	// title, after cols width
	b.SetCell(1, 1, fmt.Sprintf("%s - %s", ec.camp, ec.details))
	b.SetStyle(1, 1, sheets.Style{Bold: true, TextAlignement: sheets.ACenter})
	b.MergeCells(1, 1, 1, 4)

	// information
	b.SetCell(2, 1, "Seul le compte dédié est pris en compte.")
	b.SetStyle(2, 1, sheets.Style{Italic: true})
	b.MergeCells(2, 1, 2, 4)

	return b.Finalize()
}

// --------------------------------- bilan ---------------------------------

type campBilanVentilation struct {
	camp, asso Cent
}

// bilanCamp stores the export data with monetary amounts converted to euro cents.
type bilanCamp struct {
	Nom, Serial, Directeur, Lieu string
	Account                      ca.AccountACVE
	DateDebut                    sh.Date

	Duree, JoursInscrits, Inscrits, AuPairs, Equipiers int

	Prix, PrixJournee                  Cent
	Recettes, Remises                  Cent
	PercentageRemises, Marge           int
	DepensesTotal                      Cent
	DepenseJourneeInscrit              Cent
	fraisMDRVehicule, fraisMDRMateriel Cent
	ventilations                       [ca.VentilationCount]campBilanVentilation
}

type bilanCamps []bilanCamp

// return false if every camp has 0 in asso ventilations
func (camps bilanCamps) showAssoAccount() bool {
	for _, camp := range camps {
		if totalVentilations(camp.ventilations, false) != 0 {
			return true
		}
	}
	return false
}

// return false if every camp has 0 FraisMDR
func (camps bilanCamps) showFraisMDR() bool {
	for _, camp := range camps {
		if camp.fraisMDRVehicule != 0 || camp.fraisMDRMateriel != 0 {
			return true
		}
	}
	return false
}

func euroCent(taux ds.Taux, montant ds.Montant) Cent {
	return Cent(taux.Convertible(montant).Convert(ds.Euros).Cent)
}

func newCampBilan(stats CampSummary, taux ds.Taux) (out bilanCamp) {
	out.Nom = stats.Nom
	out.Serial = stats.Serial
	out.Directeur = stats.Directeur
	out.Lieu = stats.Lieu
	out.Account = stats.Compta.Account
	out.DateDebut = stats.DateDebut
	out.Duree = stats.Duree
	out.JoursInscrits = stats.Duree * stats.Inscrits
	out.Inscrits = stats.Inscrits
	out.AuPairs = stats.AuPairs
	out.Equipiers = stats.Equipiers
	out.Prix = euroCent(taux, stats.Prix)
	if out.Duree != 0 {
		out.PrixJournee = out.Prix / out.Duree
	}
	out.Recettes = euroCent(taux, stats.Recettes)
	out.Remises = euroCent(taux, stats.Remises)
	out.PercentageRemises = stats.PercentageRemises
	out.DepensesTotal = euroCent(taux, stats.DepensesTotal)
	out.Marge = stats.Marge
	out.DepenseJourneeInscrit = euroCent(taux, stats.DepenseJourneeInscrit)
	out.fraisMDRVehicule = stats.fraisMDRVehicule
	out.fraisMDRMateriel = stats.fraisMDRMateriel

	for v, depense := range stats.ventilations {
		camp, asso := depense.net()
		out.ventilations[v] = campBilanVentilation{
			camp: Cent(camp.AggTo(taux, ds.Euros)),
			asso: Cent(asso.AggTo(taux, ds.Euros)),
		}
	}
	return out
}

func (camp bilanCamp) dates() string {
	if camp.DateDebut.Time().IsZero() {
		return ""
	}
	debut := time.Time(camp.DateDebut)
	fin := sh.Plage{From: camp.DateDebut, Duree: camp.Duree}.To().Time()
	dM, fM := debut.Month(), fin.Month()
	if dM == fM {
		return fmt.Sprintf("du %d au %d %s", debut.Day(), fin.Day(), months[dM])
	}
	return fmt.Sprintf("du %d %s au %d %s", debut.Day(), months[dM], fin.Day(), months[fM])
}

func totalVentilations(ventilations [ca.VentilationCount]campBilanVentilation, campAccount bool) (total Cent) {
	for _, ventilation := range ventilations {
		if campAccount {
			total += ventilation.camp
		} else {
			total += ventilation.asso
		}
	}
	return total
}

// total sums additive values and recomputes ratios from the combined amounts.
func (bc bilanCamps) total() (out bilanCamp) {
	out.Nom = "Total"
	out.Serial = "ou moyenne"
	for _, camp := range bc {
		out.Duree += camp.Duree
		out.JoursInscrits += camp.JoursInscrits

		out.Inscrits += camp.Inscrits
		out.AuPairs += camp.AuPairs
		out.Equipiers += camp.Equipiers

		out.Recettes += camp.Recettes
		out.Remises += camp.Remises
		out.DepensesTotal += camp.DepensesTotal
		out.fraisMDRVehicule += camp.fraisMDRVehicule
		out.fraisMDRMateriel += camp.fraisMDRMateriel

		for v, amounts := range camp.ventilations {
			out.ventilations[v].camp += amounts.camp
			out.ventilations[v].asso += amounts.asso
		}

		out.PercentageRemises += camp.PercentageRemises
		out.Marge += camp.Marge
		out.Prix += camp.Prix
		out.PrixJournee += camp.PrixJournee
		out.DepenseJourneeInscrit += camp.DepenseJourneeInscrit
	}

	out.PercentageRemises /= len(bc)
	out.Marge /= len(bc)
	out.Prix /= len(bc)
	out.PrixJournee /= len(bc)
	out.DepenseJourneeInscrit /= len(bc)

	return out
}

func generateExcelCamps(camps bilanCamps) (*bytes.Buffer, error) {
	b := sheets.NewBuilder()

	showAssoAccount, showFraisMDR := camps.showAssoAccount(), camps.showFraisMDR()
	total := camps.total()

	const (
		color1, color2, color3, color4, color5 = "#f2f2f2", "#ffffcc", "#dae3f3", "", "#ccffcc"
		color6, color7, color8                 = color5, color5, color5
	)
	// header column
	bloc1 := [...]string{
		"Durée (en jour)",
		"Nombre de jours × Nombre d'inscrits",
		"Taille de l'équipe : au pair",
		"Taille de l'équipe : hors au pair",
		"Nombre d'inscrits",
		"Prix (pension)",
		"Prix à la journée",
	}
	bloc2 := [...]string{
		"Recettes estimées",
		"Remises",
		"Influence des remises (en %)",
	}
	bloc3 := [...]string{
		"Dépenses : Compte dédiée",
		"Dépenses : Autres comptes",
		"Dépenses : Matériel Maison du Rocher",
		"Dépenses (Total)",
	}
	bloc4 := [...]string{
		"Bénéfice",
		"Bénéfice par inscrit",
		"Coût par jour et par inscrit",
	}
	bloc5 := make([]string, ca.Bank+1)
	for v := range bloc5 {
		item := ventilationLabels[v]
		bloc5[v] = fmt.Sprintf("%s %s", item.codeACVE, item.label)
	}
	bloc6 := bloc5
	bloc7 := []string{
		"Véhicules",
		"Matériel",
	}
	if !showAssoAccount {
		bloc6 = nil
	}
	if !showFraisMDR {
		bloc7 = nil
	}

	colStart := 2
	bloc1Start := 5 + 2
	for i, title := range bloc1 {
		b.SetCell(bloc1Start+i, colStart, title)
		b.SetStyle(bloc1Start+i, colStart, sheets.Style{Color: color1, Border: sheets.Right | sheets.BorderDouble})
	}
	bloc2Start := bloc1Start + len(bloc1) + 1
	for i, title := range bloc2 {
		b.SetCell(bloc2Start+i, colStart, title)
		b.SetStyle(bloc2Start+i, colStart, sheets.Style{Color: color2, Border: sheets.Right | sheets.BorderDouble, Bold: i == 0})
	}
	bloc3Start := bloc2Start + len(bloc2) + 1
	for i, title := range bloc3 {
		b.SetCell(bloc3Start+i, colStart, title)
		b.SetStyle(bloc3Start+i, colStart, sheets.Style{Color: color3, Border: sheets.Right | sheets.BorderDouble, Bold: i == len(bloc3)-1})
	}
	bloc4Start := bloc3Start + len(bloc3) + 1
	for i, title := range bloc4 {
		b.SetCell(bloc4Start+i, colStart, title)
		b.SetStyle(bloc4Start+i, colStart, sheets.Style{Border: sheets.Right | sheets.BorderDouble, Bold: i == 2 || i == 4})
	}
	bloc5Start := bloc4Start + len(bloc4) + 1
	for i, title := range bloc5 {
		b.SetCell(bloc5Start+i, colStart, title)
		b.SetStyle(bloc5Start+i, colStart, sheets.Style{Color: color5, Border: sheets.Right | sheets.BorderDouble})
	}
	bloc6Start := bloc5Start + len(bloc5) + 1
	for i, title := range bloc6 {
		b.SetCell(bloc6Start+i, colStart, title)
		b.SetStyle(bloc6Start+i, colStart, sheets.Style{Color: color6, Border: sheets.Right | sheets.BorderDouble})
	}
	bloc7Start := bloc6Start + len(bloc6) + 1
	for i, title := range bloc7 {
		b.SetCell(bloc7Start+i, colStart, title)
		b.SetStyle(bloc7Start+i, colStart, sheets.Style{Color: color7, Border: sheets.Right | sheets.BorderDouble})
	}
	bloc8Start := bloc7Start + len(bloc7) + 1
	b.SetCell(bloc8Start, colStart, "Total")
	b.SetStyle(bloc8Start, colStart, sheets.Style{Color: color8, Border: sheets.Right | sheets.BorderDouble, Bold: true})

	// camps
	campColStart := colStart + 1
	allCamps := append(camps, total)
	for j, camp := range allCamps {
		col := campColStart + j
		// header
		account := accountLabels[camp.Account]
		dates := camp.dates()
		label := fmt.Sprintf("%s (%s)", camp.Nom, camp.Serial)
		for i, cell := range [...]string{
			label, camp.Directeur, camp.Lieu,
			account, dates,
		} {
			b.SetCell(i+1, col, cell)
			color := ""
			if i == 3 || i == 4 {
				color = accountColors[camp.Account]
			}
			var border sheets.Border
			if i == 4 {
				border = sheets.Bottom
			}
			b.SetStyle(i+1, col, sheets.Style{
				Bold: i == 0, TextAlignement: sheets.ACenter, Color: color,
				Border: border,
			})
		}

		type value struct {
			v    float32
			kind sheets.NumFormat
		}

		ventilationsCampCells, ventilationsAssoCells := make([]value, len(bloc5)), make([]value, len(bloc6))
		for v := range ventilationsCampCells {
			ventilationsCampCells[v] = value{float32(camp.ventilations[v].camp) / 100, sheets.Euros}
		}
		for v := range ventilationsAssoCells {
			ventilationsAssoCells[v] = value{float32(camp.ventilations[v].asso) / 100, sheets.Euros}
		}
		fraisMDRCells := []value{
			{float32(camp.fraisMDRVehicule) / 100, sheets.Euros},
			{float32(camp.fraisMDRMateriel) / 100, sheets.Euros},
		}
		totalDepenses := camp.DepensesTotal
		benefice := camp.Recettes - totalDepenses
		beneficeParInscrit := Cent(0)
		if camp.Inscrits != 0 {
			beneficeParInscrit = benefice / Cent(camp.Inscrits)
		}

		if !showAssoAccount {
			ventilationsAssoCells = nil
		}
		if !showFraisMDR {
			fraisMDRCells = nil
		}

		blocs := [...]struct {
			color    string
			startRow int
			cells    []value
		}{
			{
				color1, bloc1Start, []value{
					{float32(camp.Duree), sheets.Int},
					{float32(camp.JoursInscrits), sheets.Int},
					{float32(camp.AuPairs), sheets.Int},
					{float32(camp.Equipiers - camp.AuPairs), sheets.Int},
					{float32(camp.Inscrits), sheets.Int},
					{float32(camp.Prix) / 100, sheets.Euros},
					{float32(camp.PrixJournee) / 100, sheets.Euros},
				},
			},
			{
				color2, bloc2Start, []value{
					{float32(camp.Recettes) / 100, sheets.Euros},
					{float32(camp.Remises) / 100, sheets.Euros},
					{float32(camp.PercentageRemises) / 100, sheets.Percentage},
				},
			},
			{
				color3, bloc3Start, []value{
					{float32(totalVentilations(camp.ventilations, true)) / 100, sheets.Euros},
					{float32(totalVentilations(camp.ventilations, false)) / 100, sheets.Euros},
					{float32(camp.fraisMDRVehicule+camp.fraisMDRMateriel) / 100, sheets.Euros},
					{float32(totalDepenses) / 100, sheets.Euros},
				},
			},
			{
				color4, bloc4Start, []value{
					{float32(benefice) / 100, sheets.Euros},
					{float32(beneficeParInscrit) / 100, sheets.Euros},
					{float32(camp.DepenseJourneeInscrit) / 100, sheets.Euros},
				},
			},
			{
				color5, bloc5Start, ventilationsCampCells,
			},
			{
				color6, bloc6Start, ventilationsAssoCells,
			},
			{
				color7, bloc7Start, fraisMDRCells,
			},
		}
		for _, bloc := range blocs {
			for i, cell := range bloc.cells {
				if cell.v != 0 {
					b.SetCellF(bloc.startRow+i, col, cell.v)
				}
				b.SetStyle(bloc.startRow+i, col, sheets.Style{
					TextAlignement: sheets.ACenter, Color: bloc.color, Border: sheets.Right,
					NumFormat: cell.kind,
				})
			}
		}
		b.SetCellF(bloc8Start, col, float32(camp.DepensesTotal)/100)
		b.SetStyle(bloc8Start, col, sheets.Style{TextAlignement: sheets.ACenter, Color: color8, Border: sheets.Right, Bold: true, NumFormat: sheets.Euros})

	}

	b.FitColsWidth()

	// add vertical legend
	b.SetCell(bloc5Start, 1, "Compte dédié")
	b.SetStyle(bloc5Start, 1, sheets.Style{TextAlignement: sheets.AVertical})
	b.MergeCells(bloc5Start, 1, bloc5Start+len(bloc5)-1, 1)

	if showAssoAccount {
		b.SetCell(bloc6Start, 1, "Autres comptes")
		b.MergeCells(bloc6Start, 1, bloc6Start+len(bloc6)-1, 1)
		b.SetStyle(bloc6Start, 1, sheets.Style{TextAlignement: sheets.AVertical})
	}
	if showFraisMDR {
		b.SetCell(bloc7Start, 1, "Frais MDR")
		b.MergeCells(bloc7Start, 1, bloc7Start+len(bloc7)-1, 1)
		b.SetStyle(bloc7Start, 1, sheets.Style{TextAlignement: sheets.AVertical})
	}

	b.SetColumnWidth(1, 8)

	return b.Finalize()
}

// ---------------------------- Fichier excel des dépenses ----------------------------

var ventilationToColumns = [...]int{
	ca.Activites:             1,
	ca.Alimentation:          2,
	ca.FournitureEntretien:   3,
	ca.FournitureAdmin:       4,
	ca.FourniturePedagogique: 7,
	ca.Medical:               6,
	ca.Hotelier:              8,
	ca.Camping:               9,
	ca.Carburant:             5,
	ca.MaintenanceVehic:      10,
	ca.Travels:               11,
	ca.Telephone:             12,
	ca.Poste:                 13,
	ca.Bank:                  14,
	ca.NoVentilation:         15,
}

// split according to Account and sort by Date
func sortRecords(records ca.ComptaRecords, subs ca.ComptaSubrecords) (campAccount, assoAccount []ca.ComptaSubrecord) {
	for _, rec := range subs {
		if group := records[rec.IdRecord]; group.Account == ca.AccountAsso {
			assoAccount = append(assoAccount, rec)
		} else {
			campAccount = append(campAccount, rec)
		}
	}
	slices.SortFunc(campAccount, func(i, j ca.ComptaSubrecord) int { return time.Time(i.Date).Compare(time.Time(j.Date)) })
	slices.SortFunc(assoAccount, func(i, j ca.ComptaSubrecord) int { return time.Time(i.Date).Compare(time.Time(j.Date)) })
	return
}

// the files are written under the record Id
func generateExcelRecords(records ca.ComptaRecords, subrecords ca.ComptaSubrecords) (*bytes.Buffer, error) {
	b := sheets.NewBuilder()

	// header
	b.SetCell(1, 10, "Compte dédié")
	b.SetCell(1, 12, "Banque CM")
	b.SetCell(1, 15, "Ventilations (€)")
	for i, title := range [...]string{
		"", "Date", "N° pièce", "CHQ-CB-VIR", "Intitulé", "", "", "Note", "",
		"Recettes", "Dépenses", "Recettes", "Dépenses", "",
		"Activité éducative", "Alimentation", "Fourn entretien  pt équipement", "Fournitures Administratives", "Carburant", "Pharmacie Médecin", "fourniture pour activité éducative", "prestation hoteliere", "Location camping ou gite", "Maintenance véhicule", "Déplacements", "Téléphone", "Frais postaux Affranchissement", "Service bancaire", "Autre",
	} {
		b.SetCell(2, 1+i, title)
	}

	b.MergeCells(1, 10, 1, 11)
	b.MergeCells(1, 12, 1, 13)
	b.MergeCells(1, 15, 1, 31)
	b.MergeCells(2, 5, 2, 7)

	mainAccount, otherAccount := sortRecords(records, subrecords)

	writeRecords := func(title string, startRow int, recs []ca.ComptaSubrecord) {
		b.SetCell(startRow, 1, title)
		b.MergeCells(startRow, 2, startRow, 3)
		for j, record := range recs {
			row := startRow + 1 + j
			b.SetCell(row, 2, time.Time(record.Date).Format("02/01/2006"))
			if record.IdFile.Valid {
				b.SetCell(row, 3, strconv.Itoa(int(record.Id)))
			}
			b.SetCell(row, 4, record.Mode.String())
			b.SetCell(row, 5, record.Description)
			b.MergeCells(row, 5, row, 7)
			// ventilations
			for _, ventilation := range record.Montants.Uniquify() {
				offset := ventilationToColumns[ventilation.Kind]
				b.SetCell(row, 14+offset, ds.Montant{Currency: record.Currency, Cent: ventilation.Cent}.String())
			}

			group := records[record.IdRecord]
			col := 9
			if group.Account == ca.AccountAsso {
				col = 11
			}
			if !group.IsReceipe {
				col += 1
			}
			b.SetCell(row, col, "x")
			b.SetStyle(row, col, sheets.Style{Color: "#CCCCCC", TextAlignement: sheets.ACenter})
		}
	}
	// records (main account)
	writeRecords("Compte principal", 4, mainAccount)
	// records (other account)
	writeRecords("Autres comptes", 3+len(mainAccount)+3, otherAccount)

	return b.Finalize()
}
