package compta

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"registro/sql/shared"

	rd "github.com/benoitkugler/goACVE/server/core/rawdata"
	"github.com/benoitkugler/goACVE/server/core/rawdata/composites"
	"github.com/benoitkugler/goACVE/server/logs"
	"github.com/benoitkugler/goACVE/server/shared"
)

type CampHeader struct {
	IdCamp int64

	Name   string
	Year   int
	Serial string
	AtMDR  bool

	Compta ComptaExt // optional, with Id = 0 when empty

	DirecteurURL string
}

func newCampHeader(camp rd.Camp, compta ComptaCamp, records Records, host string, enc logs.Encrypteur) CampHeader {
	key, _ := shared.Encode(enc, UserKey{IsAdmin: false, IdCamp: camp.Id})
	return CampHeader{
		IdCamp: camp.Id,
		Compta: ComptaExt{
			Id:          compta.Id,
			State:       computeState(compta, records),
			Account:     compta.Account,
			RecordNotes: compta.RecordNotes,
		},
		Name:         string(camp.Nom),
		Year:         camp.DateDebut.Time().Year(),
		Serial:       serial(camp),
		AtMDR:        atMDR(string(camp.Lieu)),
		DirecteurURL: shared.BuildUrl(host, "/compta", map[string]string{"key": key}),
	}
}

func serial(camp rd.Camp) string {
	return fmt.Sprintf("CPS %d-%d", camp.Annee()-2000, camp.Id)
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
	Id          IdComptaCamp
	State       State
	Account     AccountACVE
	RecordNotes string
}

// CampDetails exposes several settings and statistics
// about one sejour.
type CampDetails struct {
	Compta ComptaExt

	Label       string
	Serial      string
	Directeur   string
	Lieu        string
	AtMDR       bool // deduce from `Lieu`
	DateDebut   Date
	DateFin     Date
	Duree       int // number of days
	Inscrits    int
	Animateurs  int
	Equipiers   int      // including animateurs/au pair
	Prix        rd.Euros // in €
	prixJournee rd.Euros

	// finances

	// total expenses, grouped by kind
	// it also includes the additional expenses
	// from FraisMDR
	ventilations ventilations

	// Computed from the pensions : amount asked + aides
	Budget float32
	// Simple estimation computed by [Prix] * [Inscrits]
	BudgetSansRemises float32
	// in %, computed from [Budget] and [BudgetSansRemises]
	PercentageRemises float32

	BudgetPrevisionnelReel   float32
	BudgetPrevisionnelSimule float32

	// Fraismdr is the total of the expenses
	// related to MDR (may be 0)
	fraisMDRVehicule float32
	fraisMDRMateriel float32 // only > 0 for camp at MDR

	// deduced from [ventilations], with remboursements deduced
	// also includes FraisMDR
	DepenseTotal, DepenseOwnAccount float32
	Marge                           float32 // in %
	DepenseJourneeInscrit           float32 // for one day and one inscrit
	Dons                            float32 // as registred by the directeur

	DepenseGraph DepenseGraph
}

func (ld *comptaLoader) compile() CampDetails {
	var out CampDetails
	out.Label = string(ld.camp.Label())
	out.Serial = serial(ld.camp)
	out.Directeur = fmt.Sprintf("%s %s", ld.directeur.FPrenom(), ld.directeur.FNom())
	out.Lieu = string(ld.camp.Lieu)
	out.AtMDR = ld.atMDR()
	out.DateDebut = Date(ld.camp.DateDebut)
	out.DateFin = Date(ld.camp.DateFin)
	out.Duree = ld.camp.Duree()
	out.Inscrits = ld.inscrits()
	out.Animateurs = len(ld.equipiers.FindAuPairs())
	out.Equipiers = len(ld.equipiers)
	out.Prix = ld.camp.Prix

	out.Compta = ComptaExt{
		Id:          ld.compta.Id,
		State:       computeState(ld.compta, ld.records),
		Account:     ld.compta.Account,
		RecordNotes: ld.compta.RecordNotes,
	}

	// finances
	fraisMDR := ld.resolveFraismdr()
	out.fraisMDRVehicule = fraisMDR.TotalBus
	if fraisMDR.HasMateriel {
		out.fraisMDRMateriel = fraisMDR.Materiel.Total
	}

	out.ventilations = ld.depenses(fraisMDR)
	out.Budget = ld.computeRecette().euros()
	out.Dons = ld.totalDons()

	out.fillFinances()

	if ld.budget.IdCompta != 0 {
		out.BudgetPrevisionnelReel = ld.resolveBudget(ld.budget, ld.config, false).Depenses
		out.BudgetPrevisionnelSimule = ld.resolveBudget(ld.budget, ld.config, true).Depenses
	}

	return out
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

func (c *CampDetails) fraisMDR() float32 { return c.fraisMDRMateriel + c.fraisMDRVehicule }

// hors MDR: 80% du budget, MDR: 50%
func (c *CampDetails) hintDepensesMax() float32 {
	if c.AtMDR {
		return c.Budget * 0.5
	}
	return c.Budget * 0.8
}

func (c *CampDetails) dates() string {
	debut, fin := time.Time(c.DateDebut), time.Time(c.DateFin)
	dM, fM := debut.Month(), fin.Month()
	if dM == fM { // condense
		return fmt.Sprintf("du %d au %d %s", debut.Day(), fin.Day(), months[dM])
	}
	return fmt.Sprintf("du %d %s au %d %s", debut.Day(), months[dM], fin.Day(), months[fM])
}

type Cent int32

func newCent(p rd.Euros) Cent { return Cent(p * 100) }

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

type comptaLoader struct {
	compta ComptaCamp
	budget ComptaBugdet // may be empty
	config Comptaconfig

	camp         rd.Camp
	factures     rd.Factures
	personnes    rd.Personnes
	participants rd.Participants
	aides        rd.Aides

	equipiers rd.Equipiers
	directeur rd.Personne // may be empty

	records    Records
	subrecords Subrecords
	forBus     RecordBuss
	dons       Doncs
}

func newComptaLoaders(db *sql.DB, ids []IdComptaCamp) ([]comptaLoader, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, shared.FormatErr("Erreur SQL", err)
	}

	comptas, err := SelectComptas(tx, ids...)
	if err != nil {
		return nil, shared.FormatErr("Erreur SQL", err)
	}
	out := make([]comptaLoader, len(ids))
	for index, id := range ids {
		out[index], err = newComptaLoader(tx, comptas[id])
		if err != nil {
			return nil, shared.FormatErr("Erreur SQL", err)
		}
	}

	_ = tx.Rollback()

	return out, nil
}

func newComptaLoader(db DB, compta ComptaCamp) (comptaLoader, error) {
	camp, err := rd.SelectCamp(db, compta.IdCamp)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	inscrits, err := rd.SelectParticipantsByIdCamps(db, camp.Id)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}
	var (
		facturesIDs  = int64Set{}
		personnesIDs = int64Set{}
	)
	for _, inscrit := range inscrits {
		if inscrit.IdFacture.IsNotNil() {
			facturesIDs.Add(inscrit.IdFacture.Int64)
		}
	}

	aides, err := rd.SelectAidesByIdParticipants(db, inscrits.Ids()...)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	factures, err := rd.SelectFactures(db, facturesIDs.Keys()...)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}
	for _, fac := range factures {
		personnesIDs.Add(fac.IdPersonne)
	}
	personnes, err := rd.SelectPersonnes(db, personnesIDs.Keys()...)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	records, subrecords, err := getRecords(db, compta.Id)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	forBus, err := SelectRecordBussByIdComptas(db, compta.Id)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	doncs, err := SelectDoncsByIdComptas(db, compta.Id)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	configs, err := SelectAllBudgetconfigs(db)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}
	campYear := camp.DateDebut.Time().Year()

	bConfig, err := selectBudgetConfig(campYear, configs)
	if err != nil {
		return comptaLoader{}, err
	}

	bugdet, _, err := SelectComptaBugdetByIdCompta(db, compta.Id)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}

	equipiers, err := rd.SelectEquipiersByIdCamps(db, compta.IdCamp)
	if err != nil {
		return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
	}
	var directeur rd.Personne
	if dir, ok := equipiers.FindDirecteur(); ok {
		directeur, err = rd.SelectPersonne(db, dir.IdPersonne)
		if err != nil {
			return comptaLoader{}, shared.FormatErr("Erreur SQL", err)
		}
	}

	return comptaLoader{compta, bugdet, bConfig, camp, factures, personnes, inscrits, aides, equipiers, directeur, records, subrecords, forBus, doncs}, nil
}

func (ld *comptaLoader) atMDR() bool { return atMDR(string(ld.camp.Lieu)) }

// do not include waiting list
func (ld *comptaLoader) inscrits() int {
	var nbInscrits int
	for _, inscrit := range ld.participants {
		if inscrit.ListeAttente.IsInscrit() {
			nbInscrits += 1
		}
	}
	return nbInscrits
}

// a.k.a. (real) budget
func (ld comptaLoader) computeRecette() Cent {
	aidesByP := map[int64][]composites.AideStructureaide{}
	for _, aide := range ld.aides {
		aidesByP[aide.Id] = append(aidesByP[aide.Id], composites.AideStructureaide{Aide: aide})
	}

	var totalAides, totalDemande rd.Euros
	for _, part := range ld.participants {
		if !part.ListeAttente.IsInscrit() {
			continue
		}

		aides := aidesByP[part.Id]
		var facture rd.Facture
		if part.IdFacture.IsNotNil() {
			facture = ld.factures[part.IdFacture.Int64]
		}
		pc := composites.ParticipantComplet{
			ParticipantCamp: composites.ParticipantCamp{Camp: ld.camp, Participant: part},
			FacturePersonne: composites.FacturePersonne{Facture: facture, Personne: ld.personnes[facture.IdPersonne]},
		}
		detailsPrix := pc.EtatFinancier(aides, false)
		totalAides += detailsPrix.TotalAides()
		totalDemande += detailsPrix.PrixNet()
	}

	return newCent(totalAides + totalDemande)
}

// agregate all records
func (ld *comptaLoader) depenses(fraisMDR ResolvedFraisMDR) ventilations {
	vs := make(ventilations)
	for _, r := range ld.subrecords {
		group := ld.records[r.IdRecord]
		vs.addMontants(r.Montants, group.IsReceipe, group.IsOtherAccount)
	}
	// add fraisMDR to other account
	vs.add(ventilations{Carburant: depenseTotale{otherAccount: fraisMDR.totalBus}})
	if fraisMDR.HasMateriel {
		vs.add(ventilations{
			FournitureEntretien:   depenseTotale{otherAccount: fraisMDR.Materiel.totalEntretien},
			FourniturePedagogique: depenseTotale{otherAccount: fraisMDR.Materiel.totalPedagogique},
		})
	}
	return vs
}

func (ld *comptaLoader) totalDons() float32 {
	var total Cent
	for _, don := range ld.dons {
		total += don.Montant
	}
	return total.euros()
}

// fill computed values
func (out *CampDetails) fillFinances() {
	out.DepenseTotal, out.DepenseOwnAccount = out.ventilations.total()

	out.prixJournee = out.Prix / rd.Euros(out.Duree)
	out.BudgetSansRemises = float32(out.Prix) * float32(out.Inscrits)
	if out.BudgetSansRemises != 0 {
		out.PercentageRemises = 100 * (out.BudgetSansRemises - out.Budget) / out.BudgetSansRemises
	}
	if di := out.Inscrits * out.Duree; di != 0 {
		out.DepenseJourneeInscrit = out.DepenseTotal / float32(di)
	}
	if out.Budget != 0 {
		out.Marge = 100 * (out.Budget - out.DepenseTotal) / out.Budget
	}
	// display both accounts
	out.DepenseGraph = out.ventilations.sortedNet()
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

// BusExpense is a summary of a [ComptaBankentry],
// keeping only the ventilation affected to cars
type BusExpense struct {
	Id    IdComptaRecord
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

func (ld *comptaLoader) resolveFraismdr() (out ResolvedFraisMDR) {
	return ld.compta.Fraismdr.resolve(ld.config, ld.subrecords, ld.forBus, ld.camp.Duree(), ld.inscrits(), len(ld.equipiers), ld.atMDR())
}

func (frais FraismdrACVE) resolve(config Comptaconfig, subrecords Subrecords, links RecordBuss, days, inscrits, equipiers int, atMDR bool) (out ResolvedFraisMDR) {
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
