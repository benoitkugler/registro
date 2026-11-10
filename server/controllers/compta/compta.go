package compta

import (
	"database/sql"
	"errors"
	"fmt"

	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	pe "registro/sql/personnes"
	sh "registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Cent = int

// CampSummary exposes several settings and statistics
// about one sejour.
type CampSummary struct {
	Compta ComptaExt

	Nom       string
	Serial    string
	Directeur string
	Lieu      string
	AtMDR     bool // deduce from `Lieu`
	DateDebut sh.Date
	Duree     int        // number of days
	Prix      ds.Montant // prix de base, sans options ou remises

	Inscrits  int // validés
	AuPairs   int // equipiers with au pair status
	Equipiers int // including animateurs/au pair

	// finances

	// Simple estimation computed by [Prix] * [Inscrits]
	Recettes ds.Montant
	// Remises is the total remises applied to the pensions
	// Note : the currency might be different from [Recettes] due to convertions
	Remises ds.Montant
	// in %, computed from [Remises] and [Recettes]
	PercentageRemises int

	BudgetPrevisionnel BudgetPrevisionnelSummary

	// MDR expenses, in euro cents.
	fraisMDRVehicule Cent
	fraisMDRMateriel Cent // only > 0 for camp at MDR

	// prixJournee rd.Euros
	// TODO
	// deduced from [ventilations], with remboursements deduced
	// also includes FraisMDR

	// total expenses, grouped by kind
	// it also includes the additional expenses
	// from FraisMDR
	ventilations ventilations

	// converted
	DepensesTotal         ds.Montant
	Marge                 int        // in %
	DepenseJourneeInscrit ds.Montant // for one day and one inscrit

	DepenseGraph DepenseGraph
}

func (ld *comptaLoader) compile() CampSummary {
	var out CampSummary

	out.Nom = ld.camp.Nom
	out.Serial = serial(ld.camp)
	out.Directeur = fmt.Sprintf("%s %s", ld.directeur.FPrenom(), ld.directeur.FNom())
	out.Lieu = ld.camp.Lieu
	out.AtMDR = ld.atMDR()
	out.DateDebut = ld.camp.DateDebut
	out.Duree = ld.camp.Duree
	out.Prix = ld.camp.Prix
	out.Inscrits = ld.inscrits()
	out.AuPairs = len(ld.equipiers.AuPairs())
	out.Equipiers = len(ld.equipiers)

	out.Compta = ComptaExt{
		Id:      ld.compta.Id,
		State:   computeState(ld.compta, ld.records),
		Account: ld.compta.Account,
	}

	// finances

	fraisMDR := ld.resolveFraismdr()
	out.fraisMDRVehicule = fraisMDR.TotalBus
	if fraisMDR.HasMateriel {
		out.fraisMDRMateriel = fraisMDR.Materiel.Total
	}
	out.ventilations = ld.aggregateDepenses(fraisMDR)

	out.Remises = ld.computeRemises()

	out.fillFinances(ld.taux)

	if ld.budget.IdCompta != 0 {
		out.BudgetPrevisionnel = BudgetPrevisionnelSummary{
			Reel:   ld.resolveBudget(ld.budget, ld.config, false).Depenses,
			Simule: ld.resolveBudget(ld.budget, ld.config, true).Depenses,
		}
	}

	return out
}

func montantToDepense(m ds.Montant, isReceipe, isAccountAsso bool) depenseTotale {
	if isReceipe {
		if isAccountAsso {
			return depenseTotale{accAssoReceipes: m.Multi()}
		}
		return depenseTotale{accCampReceipes: m.Multi()}
	} else {
		if isAccountAsso {
			return depenseTotale{accAssoDepenses: m.Multi()}
		}
		return depenseTotale{accCampDepenses: m.Multi()}
	}
}

type depenseTotale struct {
	accCampDepenses, accAssoDepenses ds.MultiCurrencies // expense, in cents
	accCampReceipes, accAssoReceipes ds.MultiCurrencies // receipe, in cents
}

// net substracts the receipe, for each account
func (dep depenseTotale) net() (camp, asso ds.MultiCurrencies) {
	return dep.accCampDepenses.Sub(dep.accCampReceipes), dep.accAssoDepenses.Sub(dep.accAssoReceipes)
}

func (dep depenseTotale) add(other depenseTotale) depenseTotale {
	return depenseTotale{
		dep.accCampDepenses.Add(other.accCampDepenses),
		dep.accAssoDepenses.Add(other.accAssoDepenses),
		dep.accCampReceipes.Add(other.accCampReceipes),
		dep.accAssoReceipes.Add(other.accAssoReceipes),
	}
}

type ventilations [ca.VentilationCount]depenseTotale

// aggregate all categories
func (vs ventilations) sum() (out depenseTotale) {
	for _, r := range vs {
		out = out.add(r)
	}
	return out
}

func (vs ventilations) total(taux ds.Taux) ds.Montant {
	s := vs.sum()
	camp, asso := s.net()
	total := camp.Add(asso)
	return total.Agg(taux)
}

func (vs *ventilations) add(other ventilations) {
	for v, m := range other {
		vs[v] = vs[v].add(m)
	}
}

func (vs *ventilations) addRecord(record ca.ComptaRecord, subrecord ca.ComptaSubrecord) {
	for _, m := range subrecord.Montants {
		vs[m.Kind] = vs[m.Kind].add(montantToDepense(ds.Montant{Cent: m.Cent, Currency: subrecord.Currency}, record.IsReceipe, record.Account == ca.AccountAsso))
	}
}

// func (vs ventilations) byAccount() (ownAccount, otherAccount map[ca.Ventilation]float32) {
// 	ownAccount, otherAccount = make(map[ca.Ventilation]float32), make(map[ca.Ventilation]float32)
// 	for v, dep := range vs {
// 		ownAccount[v] = dep.ownAccount.euros()
// 		otherAccount[v] = dep.otherAccount.euros()
// 	}
// 	return
// }

// adds both accounts
func (vs ventilations) graph(taux ds.Taux) (out DepenseGraph) {
	out.Currency = taux.Zero().Currency
	for k, v := range vs {
		camp, asso := v.net()
		m := camp.Add(asso).Agg(taux).Cent
		out.Cents[k] = m
	}
	return out
}

// DepenseGraph is the aggregated and converted sum of all expenses
type DepenseGraph struct {
	Currency ds.Currency
	Cents    [ca.VentilationCount]Cent // might contains 0
}

type VentilationMontant struct {
	Ventilation ca.Ventilation
	Montant     float32
}

type comptaLoader struct {
	camp cps.Camp
	taux ds.Taux

	compta ca.ComptaCamp
	budget ca.ComptaBugdet // may be empty
	config ca.ComptaConfig

	finances logic.DossiersFinances

	participants cps.Participants

	equipiers cps.Equipiers
	directeur pe.Personne // may be empty

	records       ca.ComptaRecords
	subrecords    ca.ComptaSubrecords
	recordsForBus ca.ComptaRecordBuss
}

func newComptaLoaders(db *sql.DB, ids []ca.IdComptaCamp) ([]comptaLoader, error) {
	out := make([]comptaLoader, len(ids))

	err := utils.InTx(db, func(tx *sql.Tx) error {
		comptas, err := ca.SelectComptaCamps(tx, ids...)
		if err != nil {
			return err
		}
		for index, id := range ids {
			out[index], err = newComptaLoader(tx, comptas[id])
			if err != nil {
				return err
			}
		}
		return nil
	})

	return out, err
}

func newComptaLoader(db ca.DB, compta ca.ComptaCamp) (comptaLoader, error) {
	camp, err := cps.SelectCamp(db, compta.IdCamp)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	taux, err := ds.SelectTaux(db, camp.IdTaux)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	bugdet, _, err := ca.SelectComptaBugdetByIdCompta(db, compta.Id)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	configs, err := ca.SelectAllComptaConfigs(db)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	campYear := camp.DateDebut.Time().Year()
	config, err := selectBudgetConfig(campYear, configs)
	if err != nil {
		return comptaLoader{}, err
	}

	participants, err := cps.SelectParticipantsByIdCamps(db, camp.Id)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	finances, err := logic.LoadDossiersFinances(db, participants.IdDossiers()...)
	if err != nil {
		return comptaLoader{}, err
	}

	records, subrecords, err := getRecords(db, compta.Id)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}

	forBus, err := ca.SelectComptaRecordBussByIdCamps(db, compta.Id)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}

	equipiers, err := cps.SelectEquipiersByIdCamps(db, compta.IdCamp)
	if err != nil {
		return comptaLoader{}, utils.SQLError(err)
	}
	var directeur pe.Personne
	if dir, ok := equipiers.Directeur(); ok {
		directeur, err = pe.SelectPersonne(db, dir.IdPersonne)
		if err != nil {
			return comptaLoader{}, utils.SQLError(err)
		}
	}

	return comptaLoader{camp, taux, compta, bugdet, config, finances, participants, equipiers, directeur, records, subrecords, forBus}, nil
}

// do not include waiting list
func (ld *comptaLoader) inscrits() int {
	var nbInscrits int
	for _, inscrit := range ld.participants {
		if inscrit.Statut == cps.Inscrit {
			nbInscrits += 1
		}
	}
	return nbInscrits
}

func (ld comptaLoader) computeRemises() ds.Montant {
	var out ds.Montant
	for id := range ld.finances.Dossiers.Dossiers {
		dossier := ld.finances.For(id)
		bilan := dossier.Bilan()
		for _, part := range dossier.Participants {
			if !(part.IdCamp == ld.camp.Id && part.Statut == cps.Inscrit) {
				continue
			}
			rem := bilan.Inscrits[part.Id].ResolvedRemises(dossier.Taux)
			// by design, all camp in a dossier have the same taux
			// so that the currency returned does not change
			out.Currency = rem.Currency
			out.Cent += rem.Cent
		}
	}
	return out
}

// agregate all records
func (ld *comptaLoader) aggregateDepenses(fraisMDR ResolvedFraisMDR) ventilations {
	var vs ventilations
	for _, r := range ld.subrecords {
		group := ld.records[r.IdRecord]
		vs.addRecord(group, r)
	}

	// add fraisMDR to other account
	vs.add(ventilations{ca.Carburant: depenseTotale{accAssoDepenses: ds.Montant{Cent: fraisMDR.TotalBus, Currency: ds.Euros}.Multi()}})
	if fraisMDR.HasMateriel {
		vs.add(ventilations{
			ca.FournitureEntretien:   depenseTotale{accAssoDepenses: ds.Montant{Cent: fraisMDR.Materiel.CostEntretien, Currency: ds.Euros}.Multi()},
			ca.FourniturePedagogique: depenseTotale{accAssoDepenses: ds.Montant{Cent: fraisMDR.Materiel.CostMateriel, Currency: ds.Euros}.Multi()},
		})
	}
	return vs
}

// fill computed values
func (out *CampSummary) fillFinances(taux ds.Taux) {
	out.Recettes = out.Prix.Mult(out.Inscrits)
	out.DepensesTotal = out.ventilations.total(taux)
	if out.Recettes.Cent != 0 {
		remises := taux.Convertible(out.Remises).Convert(out.Recettes.Currency)
		depenses := taux.Convertible(out.DepensesTotal).Convert(out.Recettes.Currency)

		out.PercentageRemises = remises.Cent * 100 / out.Recettes.Cent
		out.Marge = (out.Recettes.Cent - depenses.Cent) * 100 / out.Recettes.Cent
	}
	if di := out.Inscrits * out.Duree; di != 0 {
		out.DepenseJourneeInscrit = out.DepensesTotal.Div(di)
	}

	// display both accounts
	out.DepenseGraph = out.ventilations.graph(taux)
}

func checkVentilations(ms ca.Montants, isReceipe bool) error {
	for _, m := range ms {
		if m.Kind == ca.NoVentilation {
			return errors.New("Une ventilation doit être choisie.")
		}
		if m.Kind == ca.PensionBenevole && !isReceipe {
			return errors.New("La ventilation Pension bénévole n'est valide que pour une recette.")
		}
	}
	return nil
}

func computeState(compta ca.ComptaCamp, entries ca.ComptaRecords) ca.State {
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

type SetComptaStateIn struct {
	IdCompta ca.IdComptaCamp
	State    ca.State
	SendMail bool
}

func (ct *Controller) SetComptaState(c echo.Context) error {
	user := JWTUser(c)

	var args SetComptaStateIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.setComptaState(args, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) setComptaState(args SetComptaStateIn, user ca.IdComptaCamp) (ComptaExt, error) {
	if user != userAdmin && user != args.IdCompta {
		return ComptaExt{}, errAccess
	}

	compta, err := ca.SelectComptaCamp(ct.db, args.IdCompta)
	if err != nil {
		return ComptaExt{}, utils.SQLError(err)
	}
	camp, directorM, hasDirector, err := director(ct.db, compta)
	if err != nil {
		return ComptaExt{}, err
	}

	records, err := ca.SelectComptaRecordsByIdCamps(ct.db, args.IdCompta)
	if err != nil {
		return ComptaExt{}, utils.SQLError(err)
	}

	// ensure operation validity :
	//	- for admin : -> Archive or -> ToBeValidated
	//	- for directors : EnCours -> ToBeValidated or -> EnCours
	isValid := (user == userAdmin && (args.State == ca.Archive || args.State == ca.ToBeValidated)) ||
		(args.State == ca.EnCours || args.State == ca.ToBeValidated)
	if !isValid {
		return ComptaExt{}, fmt.Errorf("statut invalide")
	}

	compta.State = args.State
	compta, err = compta.Update(ct.db)
	if err != nil {
		return ComptaExt{}, utils.SQLError(err)
	}

	out := ComptaExt{Id: compta.Id, State: computeState(compta, records), Account: compta.Account}

	if args.SendMail {
		message := fmt.Sprintf(`
		Bonjour, <br/><br/>

		La comptabilité du séjour <i>%s</i> a évolué : le statut est maintenant <b>%s</b>. <br/><br/>

		Merci et bon courage !
		`, camp.Label(), out.State)

		if user == userAdmin && hasDirector { // notify the camp director
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(directorM, "Comptabilité des séjours", message, nil, nil)
			if err != nil {
				return ComptaExt{}, fmt.Errorf("Envoi de la notification: %s", err)
			}
		} else if user != userAdmin { // notify the compta
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(ct.adminNotificationMail, "Comptabilité des séjours", message, nil, nil)
			if err != nil {
				return ComptaExt{}, fmt.Errorf("Envoi de la notification: %s", err)
			}
		}
	}

	return out, nil
}

func (ct *Controller) GetCampDetails(c echo.Context) error {
	user := JWTUser(c)

	idCompta, err := utils.QueryParamInt[ca.IdComptaCamp](c, "IdCompta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != idCompta {
		return errAccess
	}

	out, err := ct.compileDetails(idCompta)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) compileDetails(idCompta ca.IdComptaCamp) (CampSummary, error) {
	compta, err := ca.SelectComptaCamp(ct.db, idCompta)
	if err != nil {
		return CampSummary{}, utils.SQLError(err)
	}

	ld, err := newComptaLoader(ct.db, compta)
	if err != nil {
		return CampSummary{}, err
	}

	return ld.compile(), nil
}
