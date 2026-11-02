package compta

import (
	"encoding"
	"sort"
	"strings"
	"time"
)

type Date time.Time

var (
	_ encoding.TextMarshaler   = (*Date)(nil)
	_ encoding.TextUnmarshaler = (*Date)(nil)
)

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.Time().Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	ti, err := time.Parse(time.DateOnly, string(text))
	*d = NewDateFrom(ti)
	return err
}

func newDate(year int, month time.Month, day int) Date {
	return Date(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

func NewDateFrom(ti time.Time) Date {
	return newDate(ti.Year(), ti.Month(), ti.Day())
}

func (d Date) Time() time.Time {
	ti := time.Time(d)
	return time.Date(ti.Year(), ti.Month(), ti.Day(), 0, 0, 0, 0, time.UTC)
}

// State is the global state of the compta. document
type State uint8

const (
	// The director is editing the document
	EnCours State = iota // Dossier en cours d'édition
	// The director has finished declaring entries,
	// and waits for the compta validation
	ToBeValidated // En attente de validation
	// The compta has validated every entry : every
	// new change will revert the state back to EnCours
	Valide // Dossier validé par la compta
	// The compta has closed the folder : no more changes are allowed
	// by directors
	Archive // Dossier archivé
)

func (st State) String() string {
	switch st {
	case EnCours:
		return "Dossier en cours d'édition"
	case ToBeValidated:
		return "En attente de validation"
	case Valide:
		return "Dossier validé par la compta"
	case Archive:
		return "Dossier archivé"
	default:
		panic("exhaustive switch")
	}
}

// Account maps to one of the available account names
type Account uint8

const (
	Empty Account = iota // Aucun

	MaisonFamiliale // Maison Familiale
	Clairiere       // La Clairiere
	Rocher          // Le Rocher
	Montagne        // La Montagne
	Bergerie        // La Bergerie
	Ruisseau        // Le Ruisseau
	Glandasse       // Le Glandasse
	Vercors         // Le Vercors
	Source          // La Source

	Principal      // Compte principal
	CentreVacances // Centre de vacances
	Aurore         // Aurore
)

var accountLabels = [Aurore + 1]string{
	Empty: "Aucun",

	MaisonFamiliale: "Maison Familiale",
	Clairiere:       "La Clairiere",
	Rocher:          "Le Rocher",
	Montagne:        "La Montagne",
	Bergerie:        "La Bergerie",
	Ruisseau:        "Le Ruisseau",
	Glandasse:       "Le Glandasse",
	Vercors:         "Le Vercors",
	Source:          "La Source",

	Principal:      "Compte principal",
	CentreVacances: "Centre de vacances",
	Aurore:         "Aurore",
}

var accountColors = [Aurore + 1]string{
	Empty:           "",
	MaisonFamiliale: "#c5e0b4",
	Clairiere:       "#ff6d6d",
	Rocher:          "#e575ff",
	Montagne:        "#ffcc00",
	Bergerie:        "#ffff00",
	Ruisseau:        "",
	Glandasse:       "#ffff99",
	Vercors:         "#66ff66",
	Source:          "#00ffff",
	Principal:       "",
	CentreVacances:  "",
	Aurore:          "",
}

// Reviews is a simple conversation between
// admin and director about one record
type Reviews []Review

type Review struct {
	IsAdmin bool // send by the admin ? or by the director ?
	Time    time.Time
	Message string
}

type Montants []Montant

func (ms Montants) equals(other Montants) bool {
	if len(ms) != len(other) {
		return false
	}
	for i := range ms {
		if ms[i] != other[i] {
			return false
		}
	}
	return true
}

func (ms Montants) total() Cent {
	var total Cent
	for _, m := range ms {
		total += m.Montant
	}
	return total
}

func (ms Montants) uniquify() Montants {
	tmp := map[Ventilation]Cent{}
	for _, m := range ms {
		tmp[m.Kind] += m.Montant
	}
	out := make(Montants, 0, len(tmp))
	for k, m := range tmp {
		out = append(out, Montant{m, k})
	}
	return out
}

type Montant struct {
	Montant Cent // in cents
	Kind    Ventilation
}

func (m Montant) toDepense(isReceipe, isOtherAccount bool) depenseTotale {
	if isReceipe {
		if isOtherAccount {
			return depenseTotale{otherAccountRec: m.Montant}
		}
		return depenseTotale{ownAccountRec: m.Montant}
	} else {
		if isOtherAccount {
			return depenseTotale{otherAccount: m.Montant}
		}
		return depenseTotale{ownAccount: m.Montant}
	}
}

type Ventilation uint8

const (
	NoVentilation Ventilation = iota //

	Activites             // Activités éducatives
	Alimentation          // Alimentation
	FournitureEntretien   // Fournitures d'entretien
	FournitureAdmin       // Fournitures administratives
	FourniturePedagogique // Fournitures pédagogiques
	Medical               // Pharmacie et médecins
	Hotelier              // Prestations hôtelières
	Camping               // Location de camping (ou gîte)
	Carburant             // Carburant
	MaintenanceVehic      // Maintenance des véhicules
	Travels               // Déplacements
	Telephone             // Téléphone
	Poste                 // Frais postaux
	Bank                  // Frais bancaires
	PensionBenevole       // Pension bénévole
	Autre                 // Autre
)

var ventilationLabels = [Autre + 1]struct{ code, label string }{
	NoVentilation: {"", ""},

	Activites:             {"6048", "Activités éducatives"},
	Alimentation:          {"6062", "Alimentation"},
	FournitureEntretien:   {"6063", "Fournitures d'entretien"},
	FournitureAdmin:       {"6064", "Fournitures administratives"},
	FourniturePedagogique: {"60685", "Fournitures éducatives"},
	Medical:               {"60681", "Pharmacie et médecins"},
	Hotelier:              {"6115", "Prestations hôtelières"},
	Camping:               {"613", "Location de camping (ou gîte)"},
	Carburant:             {"6066", "Carburant"},
	MaintenanceVehic:      {"6156", "Maintenance des véhicules"},
	Travels:               {"6251", "Déplacements"},
	Telephone:             {"6262", "Téléphone"},
	Poste:                 {"6263", "Frais postaux"},
	Bank:                  {"627", "Frais bancaires"},
	PensionBenevole:       {"7062", "Pension bénévole"},
	Autre:                 {"AUTRE", "Autre"},
}

func (v Ventilation) forCars() bool {
	return v == Carburant
}

func (rec *Subrecord) carExpenses() Cent {
	var total Cent
	for _, vent := range rec.Montants {
		if vent.Kind.forCars() {
			total += vent.Montant
		}
	}
	return total
}

func atMDR(lieu string) bool {
	lieu = strings.ToLower(strings.TrimSpace(lieu))
	switch lieu {
	case "chamaloc", "la maison du rocher", "maison du rocher":
		return true
	default:
		return false
	}
}

type ComptaExt struct {
	Id          IdCompta
	State       State
	Account     Account
	RecordNotes string
}

var months = [13]string{
	time.January:   "Jan.",
	time.February:  "Fév.",
	time.March:     "Mars",
	time.April:     "Avr.",
	time.May:       "Mai",
	time.June:      "Juin",
	time.July:      "Juil",
	time.August:    "Août",
	time.September: "Sep.",
	time.October:   "Oct.",
	time.November:  "Nov.",
	time.December:  "Déc.",
}

type Cent int32

func (c Cent) euros() float32 { return float32(c) / 100 }

type depenseTotale struct {
	ownAccount, otherAccount       Cent // expense, in cents
	ownAccountRec, otherAccountRec Cent // receipe, in cents
}

// net substracts the receipe, for each account
func (dep depenseTotale) net() (own, other Cent) {
	return dep.ownAccount - dep.ownAccountRec, dep.otherAccount - dep.otherAccountRec
}

func (dep depenseTotale) add(other depenseTotale) depenseTotale {
	return depenseTotale{
		dep.ownAccount + other.ownAccount,
		dep.otherAccount + other.otherAccount,
		dep.ownAccountRec + other.ownAccountRec,
		dep.otherAccountRec + other.otherAccountRec,
	}
}

type ventilations map[Ventilation]depenseTotale

// aggregate all categories
func (vs ventilations) sum() (out depenseTotale) {
	for _, r := range vs {
		out = out.add(r)
	}
	return out
}

func (vs ventilations) total() (bothAccounts, ownAccount float32) {
	s := vs.sum()
	own, other := s.net()
	return (own + other).euros(), own.euros()
}

func (vs ventilations) add(other ventilations) {
	for v, m := range other {
		vs[v] = vs[v].add(m)
	}
}

func (vs ventilations) addMontants(ms Montants, isReceipe, isOtherAccount bool) {
	for _, m := range ms {
		vs[m.Kind] = vs[m.Kind].add(m.toDepense(isReceipe, isOtherAccount))
	}
}

func (vs ventilations) byAccount() (ownAccount, otherAccount map[Ventilation]float32) {
	ownAccount, otherAccount = make(map[Ventilation]float32), make(map[Ventilation]float32)
	for v, dep := range vs {
		ownAccount[v] = dep.ownAccount.euros()
		otherAccount[v] = dep.otherAccount.euros()
	}
	return
}

// return a sorted slice, adding both accounts
func (vs ventilations) sortedNet() (out []VentilationMontant) {
	for k, v := range vs {
		own, other := v.net()
		m := own + other
		out = append(out, VentilationMontant{k, m.euros()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ventilation < out[j].Ventilation })
	return out
}

type DepenseGraph = []VentilationMontant

type VentilationMontant struct {
	Ventilation Ventilation
	Montant     float32
}

type CarKind uint8

const (
	CarBus     CarKind = iota // Bus
	CarMinibus                // Minibus
	CarCar                    // Véhicule léger
)

const nbCarKind = CarCar + 1

// CarsCosts stores the kilometric cost for each kind of cars
type CarsCosts [nbCarKind]Cent

type CarsKms [nbCarKind]int32

// Fraismdr records the indirect expenses made
// by using the material from MdR
type Fraismdr struct {
	Bus []Bus
	// PhotocopiesStart, PhotocopiesEnd int
}

type Bus struct {
	Name    string
	Kind    CarKind
	KmStart int
	KmEnd   int
}

type ResolvedBus struct {
	Bus
	Kms       int
	Tarif     float32
	TotalCost float32
}

// type ResolvedPhotocopies struct {
// 	Start, End, Number int
// 	Tarif              Cent
// 	TotalCost          float32
// }

type ResolvedMateriel struct {
	Days                                       int
	NbPeople, NbInscrits                       int
	DaysPeopleEntretien, DaysPeoplePedagogique int
	TarifEntretien, TarifPedagogique           float32
	CostEntretien, CostPedagogique             float32
	Total                                      float32

	totalEntretien, totalPedagogique Cent
}

// BusExpense is a summary of a [Record],
// keeping only the ventilation affected to cars
type BusExpense struct {
	Id    IdSubrecord
	Title string
	Total float32
}

type ResolvedFraisMDR struct {
	Bus            []ResolvedBus
	ExpensesForBus []BusExpense
	TotalBus       float32 // taking into account the potential expenses recorded
	totalBus       Cent

	// Photocopies ResolvedPhotocopies
	HasMateriel bool // true for camp at MDR
	Materiel    ResolvedMateriel

	Total float32
}

func (frais Fraismdr) resolve(config Budgetconfig, subrecords Subrecords, links RecordBuss, days, inscrits, equipiers int, atMDR bool) (out ResolvedFraisMDR) {
	out.Bus = make([]ResolvedBus, len(frais.Bus))
	for i, bus := range frais.Bus {
		tarif := config.CarsMDR[bus.Kind]
		kms := bus.KmEnd - bus.KmStart
		total := tarif * Cent(kms)
		out.Bus[i] = ResolvedBus{
			Bus:       bus,
			Kms:       kms,
			Tarif:     tarif.euros(),
			TotalCost: total.euros(),
		}
		out.totalBus += total
	}

	// select the expenses devoted to bus
	for _, link := range links {
		rec := subrecords[link.IdSubrecord]
		exp := rec.carExpenses()
		out.ExpensesForBus = append(out.ExpensesForBus, BusExpense{rec.Id, rec.Title, exp.euros()})
		out.totalBus -= exp
	}

	out.TotalBus = out.totalBus.euros()

	// photocopNb := in.PhotocopiesEnd - in.PhotocopiesStart
	// out.Photocopies = ResolvedPhotocopies{
	// 	Start: in.PhotocopiesStart, End: in.PhotocopiesEnd,
	// 	Number: photocopNb,
	// }

	all := inscrits + equipiers
	costEntretien := Cent(days*all) * config.EntretienMDR
	costMateriel := Cent(days*inscrits) * config.MaterielMDR
	out.Materiel = ResolvedMateriel{
		Days:                  days,
		NbInscrits:            inscrits,
		NbPeople:              all,
		DaysPeopleEntretien:   days * all,
		DaysPeoplePedagogique: days * inscrits,
		TarifEntretien:        config.EntretienMDR.euros(),
		TarifPedagogique:      config.MaterielMDR.euros(),
		CostEntretien:         costEntretien.euros(),
		CostPedagogique:       costMateriel.euros(),
		totalEntretien:        costEntretien,
		totalPedagogique:      costMateriel,
		Total:                 (costEntretien + costMateriel).euros(),
	}
	out.HasMateriel = atMDR

	total := out.totalBus
	if out.HasMateriel {
		total += costEntretien + costMateriel
	}
	out.Total = total.euros()

	return out
}

type FileStatus uint8

const (
	FileSystem FileStatus = iota
	GoogleDrive
	Archived
)

type OptionnalIdFilecompta struct {
	Valid bool
	Id    IdFilecompta
}

func NewOptionnalIdFilecompta(id IdFilecompta) OptionnalIdFilecompta {
	return OptionnalIdFilecompta{true, id}
}
