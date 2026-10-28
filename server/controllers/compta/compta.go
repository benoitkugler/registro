package compta

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"registro/sql/camps"
	ca "registro/sql/compta"
	"registro/sql/shared"
)

type Cent = int

type CampHeader struct {
	IdCamp int64

	Name   string
	Year   int
	Serial string
	AtMDR  bool

	Compta ComptaExt // optional, with Id = 0 when empty

	DirecteurURL string
}

func newCampHeader(camp camps.Camp, compta ca.ComptaCamp, entries ca.ComptaBankentrys, host string, enc logs.Encrypteur) CampHeader {
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

func montantToDepense(m ca.MontantPart, isReceipe, isOtherAccount bool) depenseTotale {
	if isReceipe {
		if isOtherAccount {
			return depenseTotale{otherAccountRec: m.Cent}
		}
		return depenseTotale{ownAccountRec: m.Cent}
	} else {
		if isOtherAccount {
			return depenseTotale{otherAccount: m.Cent}
		}
		return depenseTotale{ownAccount: m.Cent}
	}
}

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

type ventilations map[ca.Ventilation]depenseTotale

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

func (vs ventilations) addMontants(ms ca.Montants, isReceipe, isOtherAccount bool) {
	for _, m := range ms {
		vs[m.Kind] = vs[m.Kind].add(m.toDepense(isReceipe, isOtherAccount))
	}
}

func (vs ventilations) byAccount() (ownAccount, otherAccount map[ca.Ventilation]float32) {
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

var ventilationLabels = [ca.Autre + 1]struct{ codeACVE, label string }{
	ca.NoVentilation: {"", ""},

	ca.Activites:             {"6048", "Activités éducatives"},
	ca.Alimentation:          {"6062", "Alimentation"},
	ca.FournitureEntretien:   {"6063", "Fournitures d'entretien"},
	ca.FournitureAdmin:       {"6064", "Fournitures administratives"},
	ca.FourniturePedagogique: {"60685", "Fournitures éducatives"},
	ca.Medical:               {"60681", "Pharmacie et médecins"},
	ca.Hotelier:              {"6115", "Prestations hôtelières"},
	ca.Camping:               {"613", "Location de camping (ou gîte)"},
	ca.Carburant:             {"6066", "Carburant"},
	ca.MaintenanceVehic:      {"6156", "Maintenance des véhicules"},
	ca.Travels:               {"6251", "Déplacements"},
	ca.Telephone:             {"6262", "Téléphone"},
	ca.Poste:                 {"6263", "Frais postaux"},
	ca.Bank:                  {"627", "Frais bancaires"},
	ca.PensionBenevole:       {"7062", "Pension bénévole"},
	ca.Autre:                 {"AUTRE", "Autre"},
}

func (ms Montants) checkVentilations(isReceipe bool) error {
	for _, m := range ms {
		if m.Kind == NoVentilation {
			return errors.New("Une ventilation doit être choisie.")
		}
		if m.Kind == PensionBenevole && !isReceipe {
			return errors.New("La ventilation Pension bénévole n'est valide que pour une recette.")
		}
	}
	return nil
}

func (rec *ComptaRecord) carExpenses() Cent {
	var total Cent
	for _, vent := range rec.Montants {
		if vent.Kind == Carburant {
			total += vent.Cent
		}
	}
	return total
}

func computeState(compta ca.ComptaCamp, entries ca.ComptaBankentrys) ca.State {
	if compta.State == ca.ToBeValidated {
		// are all records valid ?
		if entries.AllValidated() {
			return ca.Valide
		}
	}
	if compta.State == ca.EnCours {
		// resolve to Valide if all records exist and are valid
		// even if the director has not ask for
		if len(entries) != 0 && entries.AllValidated() {
			return ca.Valide
		}
	}

	return compta.State
}
