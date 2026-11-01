package compta

import (
	"bytes"
	"fmt"
	"registro/generators/sheets"
	ca "registro/sql/compta"
	"registro/sql/dossiers"
	"sort"
	"strconv"
	"time"
)

// excel support

type ecritures struct {
	camp        string
	details     string
	ventilation ventilations
}

func (ld *comptaLoader) ecritures() ecritures {
	return ecritures{
		camp:        string(ld.camp.Label()),
		details:     fmt.Sprintf("%s %s - %s", ld.directeur.FPrenom(), ld.directeur.FNom(), accountLabels[ld.compta.Account]),
		ventilation: ld.depenses(ld.resolveFraismdr()),
	}
}

type ecritureComptable struct {
	NumeroCompte string
	Intitule     string
	IsCredit     bool
	Montant      float32
}

// only use the dedicated account
// toBalance is computed so that debit == credit
func (ec ecritures) compile() ([]ecritureComptable, float32) {
	// sum := ec.ventilation.sum()
	// debitTotal, receipes := sum.ownAccount, sum.ownAccountRec
	// toBalance := debitTotal - receipes

	// expList := make([]ecritureComptable, 0, len(ec.ventilation))
	// for k, dep := range ec.ventilation {
	// 	if dep.ownAccount == 0 {
	// 		continue
	// 	}
	// 	item := ventilationLabels[k]
	// 	expList = append(expList, ecritureComptable{item.codeACVE, item.label, false, dep.ownAccount.euros()})
	// }
	// sort.Slice(expList, func(i, j int) bool { return expList[i].Intitule < expList[j].Intitule })

	// // add the receipes
	// recList := make([]ecritureComptable, 0, len(ec.ventilation))
	// for k, dep := range ec.ventilation {
	// 	if dep.ownAccountRec == 0 {
	// 		continue
	// 	}
	// 	item := ventilationLabels[k]
	// 	recList = append(recList, ecritureComptable{"REC-" + item.codeACVE, "Recette - " + item.label, true, dep.ownAccountRec.euros()})
	// }
	// sort.Slice(recList, func(i, j int) bool { return recList[i].Intitule < recList[j].Intitule })

	// out := append(expList, recList...)
	// out = append([]ecritureComptable{
	// 	{"5125", ec.camp, true, toBalance.euros()},
	// }, out...)

	// return out, debitTotal.euros()
	return nil, 0
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

type bilanCamp []CampDetails

// compute the total (or average) for each field
func (bc bilanCamp) total() (out CampDetails) {
	// out.Label = "Total"
	// out.Directeur = "(ou moyenne)"
	// out.ventilations = make(ventilations)
	// for _, camp := range bc {
	// 	out.Duree += camp.Duree
	// 	out.Inscrits += camp.Inscrits
	// 	out.Animateurs += camp.Animateurs
	// 	out.Equipiers += camp.Equipiers
	// 	// out.Prix += camp.Prix

	// 	out.Budget += camp.Budget
	// 	out.BudgetSansRemises += camp.BudgetSansRemises
	// 	out.BudgetPrevisionnelReel += camp.BudgetPrevisionnelReel
	// 	out.BudgetPrevisionnelSimule += camp.BudgetPrevisionnelSimule
	// 	out.DepenseOwnAccount += camp.DepenseOwnAccount
	// 	out.DepenseTotal += camp.DepenseTotal
	// 	out.Dons += camp.Dons

	// 	out.fraisMDRVehicule += camp.fraisMDRVehicule
	// 	out.fraisMDRMateriel += camp.fraisMDRMateriel

	// 	out.prixJournee += camp.prixJournee
	// 	out.PercentageRemises += camp.PercentageRemises
	// 	out.Marge += camp.Marge
	// 	out.DepenseJourneeInscrit += camp.DepenseJourneeInscrit

	// 	// total expenses, grouped by kind
	// 	out.ventilations.add(camp.ventilations)
	// }
	// // average ratio fields
	// out.prixJournee /= rd.Euros(len(bc))
	// out.PercentageRemises /= float32(len(bc))
	// out.Marge /= float32(len(bc))
	// out.DepenseJourneeInscrit /= float32(len(bc))
	return out
}

// TODO
func generateExcelCamps(camps bilanCamp) (*bytes.Buffer, error) {
	b := sheets.NewBuilder()

	// const (
	// 	color1, color2, color3, color4, color5 = "#f2f2f2", "#ffffcc", "#dae3f3", "", "#ccffcc"
	// 	color6, color7, color8                 = color5, color5, color5
	// )
	// // header column
	// bloc1 := [...]string{
	// 	"Durée (en jour)",
	// 	"Nombre de jours × Nombre d'inscrits",
	// 	"Taille de l'équipe : animation",
	// 	"Taille de l'équipe : service",
	// 	"Nombre d'inscrits",
	// 	"Prix (pension)",
	// 	"Prix à la journée",
	// }
	// bloc2 := [...]string{
	// 	"Budget total (factures et aides)",
	// 	"Total des pensions sans remise",
	// 	"Influence des remises (en %)",
	// }
	// bloc3 := [...]string{
	// 	"Dépenses : Compte dédiée",
	// 	"Dépenses : Autres comptes",
	// 	"Dépenses : Matériel Maison du Rocher",
	// 	"Dépenses (Total)",
	// }
	// bloc4 := [...]string{
	// 	"Dépenses maximum souhaitées",
	// 	"Dons spécifiques",
	// 	"Bénéfice (hors dons)",
	// 	"Bénéfice par enfant(hors dons)",
	// 	"Coût par jour et par inscrit",
	// }
	// bloc5 := make([]string, ca.Bank+1)
	// for v := range bloc5 {
	// 	item := ventilationLabels[v]
	// 	bloc5[v] = fmt.Sprintf("%-5s %s", item.codeACVE, item.label)
	// }
	// bloc6 := bloc5
	// bloc7 := [...]string{
	// 	"Véhicules",
	// 	"Matériel",
	// }

	// colStart := 2
	// bloc1Start := 5 + 2
	// for i, title := range bloc1 {
	// 	b.SetCell(bloc1Start+i, colStart, title)
	// 	b.SetStyle(bloc1Start+i, colStart, sheets.Style{Color: color1, Border: sheets.Right | sheets.BorderDouble})
	// }
	// bloc2Start := bloc1Start + len(bloc1) + 1
	// for i, title := range bloc2 {
	// 	b.SetCell(bloc2Start+i, colStart, title)
	// 	b.SetStyle(bloc2Start+i, colStart, sheets.Style{Color: color2, Border: sheets.Right | sheets.BorderDouble, Bold: i == 0})
	// }
	// bloc3Start := bloc2Start + len(bloc2) + 1
	// for i, title := range bloc3 {
	// 	b.SetCell(bloc3Start+i, colStart, title)
	// 	b.SetStyle(bloc3Start+i, colStart, sheets.Style{Color: color3, Border: sheets.Right | sheets.BorderDouble, Bold: i == len(bloc3)-1})
	// }
	// bloc4Start := bloc3Start + len(bloc3) + 1
	// for i, title := range bloc4 {
	// 	b.SetCell(bloc4Start+i, colStart, title)
	// 	b.SetStyle(bloc4Start+i, colStart, sheets.Style{Border: sheets.Right | sheets.BorderDouble, Bold: i == 2 || i == 4})
	// }
	// bloc5Start := bloc4Start + len(bloc4) + 1
	// for i, title := range bloc5 {
	// 	b.SetCell(bloc5Start+i, colStart, title)
	// 	b.SetStyle(bloc5Start+i, colStart, sheets.Style{Color: color5, Border: sheets.Right | sheets.BorderDouble})
	// }
	// bloc6Start := bloc5Start + len(bloc5) + 1
	// for i, title := range bloc6 {
	// 	b.SetCell(bloc6Start+i, colStart, title)
	// 	b.SetStyle(bloc6Start+i, colStart, sheets.Style{Color: color6, Border: sheets.Right | sheets.BorderDouble})
	// }
	// bloc7Start := bloc6Start + len(bloc6) + 1
	// for i, title := range bloc7 {
	// 	b.SetCell(bloc7Start+i, colStart, title)
	// 	b.SetStyle(bloc7Start+i, colStart, sheets.Style{Color: color7, Border: sheets.Right | sheets.BorderDouble})
	// }
	// bloc8Start := bloc7Start + len(bloc7) + 1
	// b.SetCell(bloc8Start, colStart, "Total")
	// b.SetStyle(bloc8Start, colStart, sheets.Style{Color: color8, Border: sheets.Right | sheets.BorderDouble, Bold: true})

	// // camps
	// campColStart := colStart + 1
	// for j, camp := range append(camps, camps.total()) {
	// 	col := campColStart + j
	// 	// header
	// 	account := ""
	// 	if camp.Compta.Account != 0 {
	// 		account = accountLabels[camp.Compta.Account]
	// 	}
	// 	dates := ""
	// 	if !time.Time(camp.DateDebut).IsZero() {
	// 		dates = camp.dates()
	// 	}
	// 	for i, cell := range [...]string{
	// 		camp.Label, camp.Directeur, camp.Lieu,
	// 		account, dates,
	// 	} {
	// 		b.SetCell(i+1, col, cell)
	// 		color := ""
	// 		if i == 3 || i == 4 {
	// 			color = accountColors[camp.Compta.Account]
	// 		}
	// 		var border sheets.Border
	// 		if i == 4 {
	// 			border = sheets.Bottom
	// 		}
	// 		b.SetStyle(i+1, col, sheets.Style{
	// 			Bold: i == 0, TextAlignement: sheets.ACenter, Color: color,
	// 			Border: border,
	// 		})
	// 	}

	// 	type value struct {
	// 		v    float32
	// 		kind sheets.NumFormat
	// 	}

	// 	ventilationsOwn, ventilationsOther := camp.ventilations.byAccount()
	// 	ventilationsOwnCells, ventilationsOtherCells := make([]value, len(bloc5)), make([]value, len(bloc6))
	// 	for v := range ventilationsOwnCells {
	// 		ventilationsOwnCells[v] = value{ventilationsOwn[Ventilation(v)], sheets.Euros}
	// 	}
	// 	for v := range ventilationsOtherCells {
	// 		ventilationsOtherCells[v] = value{ventilationsOther[Ventilation(v)], sheets.Euros}
	// 	}
	// 	fraisMDR := camp.fraisMDR()
	// 	totalDepenses := camp.DepenseTotal + fraisMDR

	// 	blocs := [...]struct {
	// 		color    string
	// 		startRow int
	// 		cells    []value
	// 	}{
	// 		{
	// 			color1, bloc1Start, []value{
	// 				{float32(camp.Duree), sheets.Int},
	// 				{float32(camp.Duree * camp.Inscrits), sheets.Int},
	// 				{float32(camp.Animateurs), sheets.Int},
	// 				{float32(camp.Equipiers - camp.Animateurs), sheets.Int},
	// 				{float32(camp.Inscrits), sheets.Int},
	// 				{float32(camp.Prix), sheets.Euros},
	// 				{float32(camp.prixJournee), sheets.Euros},
	// 			},
	// 		},
	// 		{
	// 			color2, bloc2Start, []value{
	// 				{camp.Budget, sheets.Euros},
	// 				{camp.BudgetSansRemises, sheets.Euros},
	// 				{camp.PercentageRemises / 100, sheets.Percentage}, // excel % format expects a number in [0,1]
	// 			},
	// 		},
	// 		{
	// 			color3, bloc3Start, []value{
	// 				{camp.DepenseOwnAccount, sheets.Euros},
	// 				{camp.DepenseTotal - camp.DepenseOwnAccount, sheets.Euros}, // other account
	// 				{fraisMDR, sheets.Euros},
	// 				{totalDepenses, sheets.Euros},
	// 			},
	// 		},
	// 		{
	// 			color4, bloc4Start, []value{
	// 				{camp.hintDepensesMax(), sheets.Euros},
	// 				{camp.Dons, sheets.Euros},
	// 				{camp.Budget - totalDepenses, sheets.Euros},
	// 				{(camp.Budget - totalDepenses) / float32(camp.Inscrits), sheets.Euros},
	// 				{camp.DepenseJourneeInscrit, sheets.Euros},
	// 			},
	// 		},
	// 		{
	// 			color5, bloc5Start, ventilationsOwnCells,
	// 		},
	// 		{
	// 			color6, bloc6Start, ventilationsOtherCells,
	// 		},
	// 		{
	// 			color7, bloc7Start, []value{
	// 				{camp.fraisMDRVehicule, sheets.Euros},
	// 				{camp.fraisMDRMateriel, sheets.Euros},
	// 			},
	// 		},
	// 	}
	// 	for _, bloc := range blocs {
	// 		for i, cell := range bloc.cells {
	// 			if cell.v != 0 {
	// 				b.SetCellF(bloc.startRow+i, col, cell.v)
	// 			}
	// 			b.SetStyle(bloc.startRow+i, col, sheets.Style{
	// 				TextAlignement: sheets.ACenter, Color: bloc.color, Border: sheets.Right,
	// 				NumFormat: cell.kind,
	// 			})
	// 		}
	// 	}
	// 	b.SetCellF(bloc8Start, col, camp.DepenseTotal)
	// 	b.SetStyle(bloc8Start, col, sheets.Style{TextAlignement: sheets.ACenter, Color: color8, Border: sheets.Right, Bold: true})

	// }

	// b.FitColsWidth()

	// // add vertical legend
	// b.SetCell(bloc5Start, 1, "Compte dédié")
	// b.SetStyle(bloc5Start, 1, sheets.Style{TextAlignement: sheets.AVertical})
	// b.MergeCells(bloc5Start, 1, bloc5Start+len(bloc5)-1, 1)
	// b.SetCell(bloc6Start, 1, "Autres comptes")
	// b.MergeCells(bloc6Start, 1, bloc6Start+len(bloc6)-1, 1)
	// b.SetStyle(bloc6Start, 1, sheets.Style{TextAlignement: sheets.AVertical})
	// b.SetCell(bloc7Start, 1, "Frais MDR")
	// b.MergeCells(bloc7Start, 1, bloc7Start+len(bloc7)-1, 1)
	// b.SetStyle(bloc7Start, 1, sheets.Style{TextAlignement: sheets.AVertical})

	// b.SetColumnWidth(1, 8)

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

// split according to IsOtherAccount and sort by Date
func sortRecords(records ca.ComptaRecords, subs ca.ComptaSubrecords) (mainAccount, otherAccount []ca.ComptaSubrecord) {
	for _, rec := range subs {
		if group := records[rec.IdRecord]; group.Account == ca.AccountAsso {
			otherAccount = append(otherAccount, rec)
		} else {
			mainAccount = append(mainAccount, rec)
		}
	}
	sort.Slice(mainAccount, func(i, j int) bool { return time.Time(mainAccount[i].Date).Before(time.Time(mainAccount[j].Date)) })
	sort.Slice(otherAccount, func(i, j int) bool { return time.Time(otherAccount[i].Date).Before(time.Time(otherAccount[j].Date)) })
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
				b.SetCell(row, 14+offset, dossiers.Montant{Currency: record.Currency, Cent: ventilation.Cent}.String())
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
