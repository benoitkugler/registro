package compta

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"time"

	"registro/crypto"
	"registro/logic"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	sh "registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type CampHeader struct {
	logic.CampItem

	Serial string
	AtMDR  bool

	Compta ComptaExt // optional, with Id = 0 when empty

	DirecteurURL string
}

func newCampHeader(camp cps.Camp, compta ca.ComptaCamp, entries ca.ComptaRecords, host string, enc crypto.Encrypter) CampHeader {
	return CampHeader{
		CampItem: logic.NewCampItem(camp),
		Serial:   serial(camp),
		AtMDR:    atMDR(string(camp.Lieu)),
		Compta: ComptaExt{
			compta.Id,
			computeState(compta, entries),
			compta.Account,
		},
		DirecteurURL: DirecteurURL(enc, host, camp.Id),
	}
}

func serial(camp cps.Camp) string {
	return fmt.Sprintf("CPS %d-%d", camp.DateDebut.Time().Year()-2000, camp.Id)
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

func (c *CampSummary) dates() string {
	debut, fin := time.Time(c.DateDebut), sh.Plage{From: c.DateDebut, Duree: c.Duree}.To().Time()
	dM, fM := debut.Month(), fin.Month()
	if dM == fM { // condense
		return fmt.Sprintf("du %d au %d %s", debut.Day(), fin.Day(), months[dM])
	}
	return fmt.Sprintf("du %d %s au %d %s", debut.Day(), months[dM], fin.Day(), months[fM])
}

type ComptaExt struct {
	Id      ca.IdComptaCamp
	State   ca.State
	Account ca.AccountACVE
}

// GetCamps returns the list of existing camps, even
// the ones for which no compta entry exist yet.
// Admin only
func (ct *Controller) GetCamps(c echo.Context) error {
	user := JWTUser(c)
	if user != userAdmin {
		return errAccess
	}

	out, err := ct.getCamps(c.Request().Host)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) getCamps(host string) (out []CampHeader, _ error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	comptas, err := ca.SelectAllComptaCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	byCamp := comptas.ByIdCamp()

	records, err := ca.SelectAllComptaRecords(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	byCompta := records.ByIdCamp()

	for _, camp := range camps {
		compta := byCamp[camp.Id]
		records := byCompta[compta.Id]

		out = append(out, newCampHeader(camp, compta, records, host, ct.key))
	}

	// recent first
	slices.SortFunc(out, func(a, b CampHeader) int { return cmp.Compare(b.DateDebut.Time().Year(), a.DateDebut.Time().Year()) })
	return out, nil
}

func (ct *Controller) GetCompta(c echo.Context) error {
	user := JWTUser(c)

	idCompta, err := utils.QueryParamInt[ca.IdComptaCamp](c, "IdCompta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != idCompta {
		return errAccess
	}

	out, err := ct.getCompta(c.Request().Host, idCompta)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) getCompta(host string, id ca.IdComptaCamp) (CampHeader, error) {
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}

	camp, err := cps.SelectCamp(ct.db, compta.IdCamp)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}

	records, err := ca.SelectComptaRecordsByIdCamps(ct.db, id)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}

	return newCampHeader(camp, compta, records, host, ct.key), nil
}

func (ct *Controller) CreateCompta(c echo.Context) error {
	user := JWTUser(c)

	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "IdCamp")
	if err != nil {
		return err
	}

	if user != userAdmin {
		return errAccess
	}

	co, err := ct.loadOrCreateCompta(idCamp)
	if err != nil {
		return err
	}

	return c.JSON(200, co)
}

func (ct *Controller) DeleteCompta(c echo.Context) error {
	user := JWTUser(c)

	idC, err := utils.QueryParamInt[ca.IdComptaCamp](c, "IdCompta")
	if err != nil {
		return err
	}

	if user != userAdmin {
		return errAccess
	}

	err = ct.deleteCompta(idC)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) deleteCompta(id ca.IdComptaCamp) error {
	// delete documents
	_, subrecords, err := getRecords(ct.db, id)
	if err != nil {
		return err
	}
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}

	metas, err := ca.SelectComptaFiles(ct.db, subrecords.IdFiles()...)
	if err != nil {
		return utils.SQLError(err)
	}

	go func() {
		err := ct.files.removeMany(ct.db, compta.FileStatus, metas)
		if err != nil {
			log.Println(err)
		}
	}()

	// SQL items will cascade
	_, err = ca.DeleteComptaCampById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}

	return nil
}

func (ct *Controller) UpdateCompta(c echo.Context) error {
	user := JWTUser(c)

	var args ComptaExt
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.Id {
		return errAccess
	}

	err := ct.updateCompta(args)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) updateCompta(compta ComptaExt) error {
	existing, err := ca.SelectComptaCamp(ct.db, compta.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	existing.Account = compta.Account
	_, err = existing.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}
