package compta

import (
	"database/sql"
	"strings"
	"time"

	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	sh "registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// custom logic for ACVE

func (ld *comptaLoader) atMDR() bool { return atMDR(string(ld.camp.Lieu)) }

func atMDR(lieu string) bool {
	lieu = strings.ToLower(strings.TrimSpace(lieu))
	switch lieu {
	case "chamaloc", "la maison du rocher", "maison du rocher":
		return true
	default:
		return false
	}
}

var accountLabels = [ca.Aurore + 1]string{
	ca.Empty: "Aucun",

	ca.MaisonFamiliale: "Maison Familiale",
	ca.Clairiere:       "La Clairiere",
	ca.Rocher:          "Le Rocher",
	ca.Montagne:        "La Montagne",
	ca.Bergerie:        "La Bergerie",
	ca.Ruisseau:        "Le Ruisseau",
	ca.Glandasse:       "Le Glandasse",
	ca.Vercors:         "Le Vercors",
	ca.Source:          "La Source",

	ca.Principal:      "Compte principal",
	ca.CentreVacances: "Centre de vacances",
	ca.Aurore:         "Aurore",
}

var accountColors = [ca.Aurore + 1]string{
	ca.Empty:           "",
	ca.MaisonFamiliale: "#c5e0b4",
	ca.Clairiere:       "#ff6d6d",
	ca.Rocher:          "#e575ff",
	ca.Montagne:        "#ffcc00",
	ca.Bergerie:        "#ffff00",
	ca.Ruisseau:        "",
	ca.Glandasse:       "#ffff99",
	ca.Vercors:         "#66ff66",
	ca.Source:          "#00ffff",
	ca.Principal:       "",
	ca.CentreVacances:  "",
	ca.Aurore:          "",
}

func (ld *comptaLoader) budgetFraisMDR(config ca.ComptaConfig, totalPeople int) (chauffage, entretien ds.Montant) {
	// chauffage is activated from november to april
	m := ld.camp.DateDebut.Time().Month()
	if m >= time.November || m <= time.April {
		chauffage = ds.Montant{Cent: config.ChauffageMDR * ld.camp.Duree, Currency: ds.Euros}
	}
	// number of days * number of people
	daysPeople := ld.camp.Duree * totalPeople
	entretien = ds.Montant{Cent: config.EntretienMDR * daysPeople, Currency: ds.Euros}

	return
}

// only use [ca.Carburant] ventilations
func carExpenses(rec *ca.ComptaSubrecord) Cent {
	var total Cent
	for _, vent := range rec.Montants {
		if vent.Kind == ca.Carburant {
			total += vent.Cent
		}
	}
	return total
}

type ResolvedBus struct {
	ca.Bus
	Kms       int
	Tarif     ds.Montant
	TotalCost ds.Montant
}

type ResolvedMateriel struct {
	Days                                       int
	NbPeople, NbInscrits                       int
	DaysPeopleEntretien, DaysPeoplePedagogique int

	TarifEntretien, TarifMateriel Cent
	CostEntretien, CostMateriel   Cent // computed from days, people and tarif

	Total Cent
}

// BusExpense is a summary of a [ComptaRecord],
// keeping only the ventilation affected to cars
type BusExpense struct {
	Id          ca.IdComptaSubrecord
	Description string
	Total       ds.Montant
}

// all value are in €, since it is only used
// for ACVE
type ResolvedFraisMDR struct {
	Bus            []ResolvedBus
	ExpensesForBus []BusExpense
	TotalBus       Cent // taking into account the potential expenses recorded

	HasMateriel bool // true for camp at MDR
	Materiel    ResolvedMateriel

	Total Cent
}

func (ld *comptaLoader) resolveFraisMDR() (out ResolvedFraisMDR) {
	return resolveFraismdr(ld.compta.FraisMDR, ld.config, ld.subrecords, ld.recordsForBus, ld.camp.Duree, ld.inscrits(), len(ld.equipiers), ld.atMDR())
}

func resolveFraismdr(frais ca.FraisMDRACVE, config ca.ComptaConfig, subrecords ca.ComptaSubrecords, links ca.ComptaRecordBuss, days, inscrits, equipiers int, atMDR bool) (out ResolvedFraisMDR) {
	out.Bus = make([]ResolvedBus, len(frais.Bus))
	for i, bus := range frais.Bus {
		tarif := ds.Montant{Cent: Cent(config.AssoCars[bus.Kind]), Currency: ds.Euros}
		kms := bus.KmEnd - bus.KmStart
		total := tarif.Mult(kms)
		out.Bus[i] = ResolvedBus{
			Bus:       bus,
			Kms:       kms,
			Tarif:     tarif,
			TotalCost: total,
		}
		out.TotalBus += total.Cent
	}

	// select the expenses devoted to bus
	for _, link := range links {
		rec := subrecords[link.IdSubrecord]
		exp := carExpenses(&rec)
		out.ExpensesForBus = append(out.ExpensesForBus, BusExpense{rec.Id, rec.Description, ds.Montant{Cent: exp, Currency: ds.Euros}})
		out.TotalBus -= exp
	}

	out.HasMateriel = atMDR
	all := inscrits + equipiers
	costEntretien := days * all * config.EntretienMDR
	costMateriel := days * inscrits * config.MaterielMDR
	out.Materiel = ResolvedMateriel{
		Days:                  days,
		NbPeople:              all,
		NbInscrits:            inscrits,
		DaysPeopleEntretien:   days * all,
		DaysPeoplePedagogique: days * inscrits,
		TarifEntretien:        config.EntretienMDR,
		TarifMateriel:         config.MaterielMDR,
		CostEntretien:         costEntretien,
		CostMateriel:          costMateriel,
		Total:                 costEntretien + costMateriel,
	}

	total := out.TotalBus
	if out.HasMateriel {
		total += costEntretien + costMateriel
	}
	out.Total = total

	return out
}

func (ct *Controller) FraisMDRGetBusExpense(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != id {
		return errAccess
	}

	out, err := ct.loadBusExpenses(id)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type RecordBusHeader struct {
	IdSubrecord ca.IdComptaSubrecord
	Title       string
	Date        sh.Date
	ForBus      ds.Montant
	Total       ds.Montant
}

type BusExpensesOut struct {
	Records  []RecordBusHeader // only the one with car ventilation
	Selected []ca.IdComptaSubrecord
}

func (ct *Controller) loadBusExpenses(id ca.IdComptaCamp) (BusExpensesOut, error) {
	records, subs, err := getRecords(ct.db, id)
	if err != nil {
		return BusExpensesOut{}, utils.SQLError(err)
	}

	forBus, err := ca.SelectComptaRecordBussByIdCamps(ct.db, id)
	if err != nil {
		return BusExpensesOut{}, utils.SQLError(err)
	}

	var candidates []RecordBusHeader
	for _, rec := range subs {
		carExp := carExpenses(&rec)
		if group := records[rec.IdRecord]; group.IsReceipe || carExp == 0 {
			continue
		}
		candidates = append(candidates, RecordBusHeader{
			IdSubrecord: rec.Id,
			Title:       rec.Description,
			Date:        rec.Date,
			ForBus:      ds.Montant{Cent: carExp, Currency: ds.Euros},
			Total:       ds.Montant{Cent: rec.Montants.Total(), Currency: ds.Euros},
		})
	}

	return BusExpensesOut{Records: candidates, Selected: forBus.IdSubrecords()}, nil
}

type SetBusExpenseIn struct {
	IdCompta ca.IdComptaCamp
	Records  []ca.IdComptaSubrecord
}

func (ct *Controller) FraisMDRSetBusExpense(c echo.Context) error {
	user := JWTUser(c)

	var args SetBusExpenseIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.IdCompta {
		return errAccess
	}

	err := ct.setBusExpense(args)
	if err != nil {
		return err
	}

	out, err := ct.loadFraisMDR(args.IdCompta)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) setBusExpense(args SetBusExpenseIn) error {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err := ca.DeleteComptaRecordBussByIdCamps(tx, args.IdCompta)
		if err != nil {
			return err
		}

		links := make(ca.ComptaRecordBuss, len(args.Records))
		for i, id := range args.Records {
			links[i] = ca.ComptaRecordBus{IdCamp: args.IdCompta, IdSubrecord: id}
		}
		err = ca.InsertManyComptaRecordBuss(tx, links...)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (ct *Controller) FraisMDRLoad(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != id {
		return errAccess
	}

	out, err := ct.loadFraisMDR(id)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) loadFraisMDR(id ca.IdComptaCamp) (ResolvedFraisMDR, error) {
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return ResolvedFraisMDR{}, utils.SQLError(err)
	}

	ld, err := newComptaLoader(ct.db, compta)
	if err != nil {
		return ResolvedFraisMDR{}, err
	}

	return ld.resolveFraisMDR(), nil
}

type UpdateFraisMDRIn struct {
	IdCompta ca.IdComptaCamp
	Frais    ca.FraisMDRACVE
}

func (ct *Controller) FraisMDRUpdate(c echo.Context) error {
	user := JWTUser(c)

	var args UpdateFraisMDRIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.IdCompta {
		return errAccess
	}

	compta, err := ca.SelectComptaCamp(ct.db, args.IdCompta)
	if err != nil {
		return utils.SQLError(err)
	}
	compta.FraisMDR = args.Frais
	_, err = compta.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}

	out, err := ct.loadFraisMDR(args.IdCompta)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}
