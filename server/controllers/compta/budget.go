package compta

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"registro/logic"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type BudgetPrevisionnelSummary struct {
	Reel   ds.Montant
	Simule ds.Montant
}

// returns the more recent configuration among the ones available at the sejour time
func selectBudgetConfig(campYear int, configs ca.ComptaConfigs) (ca.ComptaConfig, error) {
	sorted := configs.Sorted()
	// find the best match, that is the first year before (or equal) [campYear]
	for _, config := range sorted {
		if config.Year <= campYear { // got it
			return config, nil
		}
	}
	return ca.ComptaConfig{}, fmt.Errorf("internal error: aucune configuration n'est disponible pour l'année %d", campYear)
}

// ResolvedBudget exposes the dynamic fields, computed
// using camp statistics
type ResolvedBudget struct {
	Duree     int // days
	Inscrits  int
	Equipiers int
	Prix      ds.Montant

	FoodInscrits       ds.Montant
	FoodEquipiers      ds.Montant
	HasSmallTeamFactor bool
	FoodTotal          ds.Montant

	ChauffageMDR ds.Montant
	EntretienMDR ds.Montant
	HostTotal    ds.Montant // external or MDR (if any)

	MaterielMDR     ds.Montant
	ActivitiesTotal ds.Montant

	TransportTotal ds.Montant

	Pensions        ds.Montant
	Depenses        ds.Montant
	Bilan           ds.Montant // pensions - depenses
	BilanPercentage int
}

// if [simulated] is true, the number of inscrits and equipiers are taken from the budget field
// otherwise they are computed from the database
func (ld *comptaLoader) resolveBudget(origin ca.ComptaBudget, config ca.ComptaConfig, simulated bool) ResolvedBudget {
	out := ResolvedBudget{}

	out.Duree = ld.camp.Duree
	if simulated {
		out.Inscrits = origin.SimulatedInscrits
		out.Equipiers = origin.SimulatedEquipiers
	} else {
		out.Inscrits = ld.inscrits()
		out.Equipiers = len(ld.equipiers)
	}
	out.Prix = ld.camp.Prix

	foodInscrits := resolveMealsCost(config.MealsCost, out.Inscrits, resolveMealsCount(origin.StartMealInscrits, origin.EndMealInscrits, out.Duree, 0, 0))
	foodEquipiers := resolveMealsCost(config.MealsCost, out.Equipiers, resolveMealsCount(origin.StartMealEquipiers, origin.EndMealEquipiers, out.Duree, origin.OffsetStartMealEquipiers, origin.OffsetEndMealEquipiers))
	out.HasSmallTeamFactor = out.Inscrits < config.SmallMealThreshold
	percent := 0
	if out.HasSmallTeamFactor {
		percent = config.SmallMealMajoration
	}
	out.FoodInscrits = ds.Montant{Cent: foodInscrits, Currency: ds.Euros}
	out.FoodEquipiers = ds.Montant{Cent: foodEquipiers, Currency: ds.Euros}
	out.FoodTotal = ds.Montant{Cent: (foodInscrits + foodEquipiers) * (100 + percent) / 100, Currency: ds.Euros}

	atMDR := ld.atMDR()
	if atMDR {
		out.ChauffageMDR, out.EntretienMDR = ld.budgetFraisMDR(config, out.Inscrits+out.Equipiers)
		// host cost is only MDR
		out.HostTotal = ds.Montant{Cent: out.ChauffageMDR.Cent + out.EntretienMDR.Cent, Currency: ds.Euros}
	} else {
		// host cost is computed with user provided data
		out.HostTotal = resolveCostList(ld.taux, origin.ExternalHost)
	}

	var materielMDR ds.Montant
	if atMDR {
		materielMDR = ds.Montant{Cent: config.MaterielMDR * out.Inscrits * out.Duree, Currency: ds.Euros}
	}
	// total activities
	activitiesTotal := ld.taux.Zero()
	activitiesTotal.Add(materielMDR)
	activitiesTotal.Add(resolveCostList(ld.taux, origin.Activities))
	activitiesTotal.Add(origin.MaterielPeda)
	activitiesTotal.Add(origin.ProduitPharma)

	transportTotal := ld.taux.Zero()
	transportTotal.Add(origin.TeamTrips)
	transportTotal.Add(origin.ExternalCarRent)
	transportTotal.Add(origin.AssoCarKms.Resolve(config.AssoCars))

	out.MaterielMDR = materielMDR
	out.ActivitiesTotal = activitiesTotal.Montant
	out.TransportTotal = transportTotal.Montant

	depenses := ld.taux.Zero()
	depenses.Add(out.FoodTotal)
	depenses.Add(out.HostTotal)
	depenses.Add(out.ActivitiesTotal)
	depenses.Add(out.TransportTotal)

	out.Pensions = out.Prix.Mult(out.Inscrits)
	out.Depenses = depenses.Montant
	depensesC := ld.taux.Convertible(out.Depenses).Convert(out.Pensions.Currency)
	out.Bilan = ds.Montant{Cent: out.Pensions.Cent - depensesC.Cent, Currency: out.Pensions.Currency}
	if out.Pensions.Cent != 0 {
		out.BilanPercentage = 100 * out.Bilan.Cent / out.Pensions.Cent
	}

	return out
}

// Number of meals during the whole vacation, for ONE person
type mealsCount [ca.Cinquieme + 1]int

func resolveMealsCount(startMeal, endMeal ca.Meal, duree, startOffset, endOffset int) mealsCount {
	// start by full days, then adjust
	duree += (endOffset - startOffset)
	out := mealsCount{duree, duree, duree, duree, duree}
	for i := range startMeal { // remove one meal before start (if any)
		out[i] -= 1
	}
	for i := endMeal + 1; i < ca.Cinquieme+1; i++ { // remove one meal after end (if any)
		out[i] -= 1
	}
	return out
}

// in € cents
func resolveMealsCost(mc ca.MealsCosts, people int, meals mealsCount) Cent {
	var total Cent
	for i, cost := range mc {
		count := meals[i]
		total += Cent(cost) * Cent(count)
	}
	return total * people
}

func resolveCostList(taux ds.Taux, costs []ca.NamedCost) ds.Montant {
	out := taux.Zero()
	for _, c := range costs {
		out.Add(c.Cost)
	}
	return out.Montant
}

// API

func (ct *Controller) GetComptaConfigs(c echo.Context) error {
	user := JWTUser(c)

	if user != userAdmin {
		return errAccess
	}

	configs, err := ca.SelectAllComptaConfigs(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	out := configs.Sorted()

	return c.JSON(200, out)
}

func (ct *Controller) CreateComptaConfig(c echo.Context) error {
	user := JWTUser(c)

	if user != userAdmin {
		return errAccess
	}

	out, err := ct.createConfig()
	if err != nil {
		return utils.SQLError(err)
	}

	return c.JSON(200, out)
}

func (ct *Controller) createConfig() (ca.ComptaConfig, error) {
	// copy the latest one, if any
	configs, err := ca.SelectAllComptaConfigs(ct.db)
	if err != nil {
		return ca.ComptaConfig{}, utils.SQLError(err)
	}
	var config ca.ComptaConfig
	if l := configs.Sorted(); len(l) != 0 {
		config = l[0]
	}

	config.Year = time.Now().Year()

	out, err := config.Insert(ct.db)
	if err != nil {
		return ca.ComptaConfig{}, utils.SQLError(err)
	}
	return out, nil
}

func (ct *Controller) UpdateComptaConfig(c echo.Context) error {
	user := JWTUser(c)

	if user != userAdmin {
		return errAccess
	}

	var args ca.ComptaConfig
	if err := c.Bind(&args); err != nil {
		return err
	}
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}

	return c.NoContent(200)
}

func (ct *Controller) DeleteComptaConfig(c echo.Context) error {
	user := JWTUser(c)

	if user != userAdmin {
		return errAccess
	}

	id, err := utils.QueryParamInt[ca.IdComptaConfig](c, "id")
	if err != nil {
		return err
	}
	_, err = ca.DeleteComptaConfigById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}

	return c.NoContent(200)
}

type CreateBudgetIn struct {
	IdCompta   ca.IdComptaCamp
	ImportFrom ca.IdComptaCamp // 0 to create a blank budget
}

func (ct *Controller) CreateBudget(c echo.Context) error {
	user := JWTUser(c)

	var args CreateBudgetIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.IdCompta {
		return errAccess
	}

	err := ct.createBudget(args)
	if err != nil {
		return err
	}

	out, err := ct.loadEstimatedBudget(args.IdCompta)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) createBudget(args CreateBudgetIn) (err error) {
	var newBudget ca.ComptaBudget
	if args.ImportFrom != 0 {
		newBudget, _, err = ca.SelectComptaBudgetByIdCompta(ct.db, args.ImportFrom)
		if err != nil {
			return utils.SQLError(err)
		}
	}
	newBudget.IdCompta = args.IdCompta

	err = newBudget.Insert(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) GetEstimatedBudget(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != id {
		return errAccess
	}

	out, err := ct.loadEstimatedBudget(id)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

// EstimatedBudget may be empty if not filled yet
type EstimatedBudget struct {
	IsValid bool
	Origin  ca.ComptaBudget
	Config  ca.ComptaConfig

	Simulated ResolvedBudget
	Real      ResolvedBudget
}

func (ct *Controller) loadEstimatedBudget(idCompta ca.IdComptaCamp) (EstimatedBudget, error) {
	compta, err := ca.SelectComptaCamp(ct.db, idCompta)
	if err != nil {
		return EstimatedBudget{}, utils.SQLError(err)
	}
	ld, err := newComptaLoader(ct.db, compta)
	if err != nil {
		return EstimatedBudget{}, err
	}

	if ld.budget.IdCompta == 0 {
		return EstimatedBudget{IsValid: false}, nil
	}

	return EstimatedBudget{
		IsValid:   true,
		Origin:    ld.budget,
		Config:    ld.config,
		Simulated: ld.resolveBudget(ld.budget, ld.config, true),
		Real:      ld.resolveBudget(ld.budget, ld.config, false),
	}, nil
}

func (ct *Controller) UpdateEstimatedBudget(c echo.Context) error {
	user := JWTUser(c)

	var args ca.ComptaBudget
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.IdCompta {
		return errAccess
	}

	out, err := ct.updateBudget(args)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) updateBudget(origin ca.ComptaBudget) (EstimatedBudget, error) {
	err := origin.Delete(ct.db)
	if err != nil {
		return EstimatedBudget{}, utils.SQLError(err)
	}

	err = origin.Insert(ct.db)
	if err != nil {
		return EstimatedBudget{}, utils.SQLError(err)
	}

	return ct.loadEstimatedBudget(origin.IdCompta)
}

type CampBudgetHeader struct {
	IdCompta ca.IdComptaCamp
	logic.CampItem
}

func (ct *Controller) GetPreviousBudgets(c echo.Context) error {
	out, err := ct.getPreviousBudgets()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadCampsByBudget(budgets ca.ComptaBudgets) (map[ca.IdComptaCamp]CampBudgetHeader, error) {
	comptas, err := ca.SelectComptaCamps(ct.db, budgets.IdComptas()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	byCamp := comptas.ByIdCamp()

	camps, err := cps.SelectCamps(ct.db, comptas.IdCamps()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	out := make(map[ca.IdComptaCamp]CampBudgetHeader)
	for _, camp := range camps {
		idCompta := byCamp[camp.Id].Id
		out[idCompta] = CampBudgetHeader{
			IdCompta: idCompta,
			CampItem: logic.NewCampItem(camp),
		}
	}

	return out, nil
}

func (ct *Controller) getPreviousBudgets() ([]CampBudgetHeader, error) {
	budgets, err := ca.SelectAllComptaBudgets(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	headers, err := ct.loadCampsByBudget(budgets)
	if err != nil {
		return nil, err
	}

	var out []CampBudgetHeader
	for _, camp := range headers {
		out = append(out, camp)
	}

	slices.SortFunc(out, func(a, b CampBudgetHeader) int {
		if yearOrder := cmp.Compare(a.DateDebut.Time().Year(), b.DateDebut.Time().Year()); yearOrder != 0 {
			return yearOrder
		}
		return strings.Compare(a.Nom, b.Nom)
	})

	return out, nil
}

// GetBudgetActivities returns all the activities declared so far,
// expected the one in the current budget.
func (ct *Controller) GetBudgetActivities(c echo.Context) error {
	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	out, err := ct.budgetActivities(id)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type ActivityHint struct {
	Activity ca.NamedCost
	Origin   CampBudgetHeader
}

func (ct *Controller) budgetActivities(id ca.IdComptaCamp) ([]ActivityHint, error) {
	budgets, err := ca.SelectAllComptaBudgets(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	camps, err := ct.loadCampsByBudget(budgets)
	if err != nil {
		return nil, err
	}

	type key struct {
		act  string
		camp string
	}
	uniq := map[key]ActivityHint{}
	seen := map[ca.NamedCost]bool{}
	for _, budget := range budgets {
		if budget.IdCompta == id {
			continue
		}

		camp := camps[budget.IdCompta]
		for _, activity := range budget.Activities {
			// de-duplicate exact activities
			if seen[activity] {
				continue
			}

			k := key{activity.Name, camp.Nom}
			// only keep more recent version
			if other, has := uniq[k]; has && other.Origin.DateDebut.Time().Year() > camp.DateDebut.Time().Year() {
				continue
			}
			uniq[k] = ActivityHint{Activity: activity, Origin: camp}
			seen[activity] = true
		}
	}

	var out []ActivityHint
	for _, v := range uniq {
		out = append(out, v)
	}

	slices.SortFunc(out, func(a, b ActivityHint) int {
		if nameOrder := strings.Compare(a.Activity.Name, b.Activity.Name); nameOrder != 0 {
			return nameOrder
		}
		return cmp.Compare(b.Origin.DateDebut.Time().Year(), a.Origin.DateDebut.Time().Year())
	})

	return out, nil
}
