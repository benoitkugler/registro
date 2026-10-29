package compta

import (
	"fmt"
	"registro/sql/camps"
	ca "registro/sql/compta"
	"registro/utils"
)

func (ct *Controller) getRecordsArchive(id ca.IdComptaCamp) (content []byte, filename string, _ error) {
	recs, subs, err := getRecords(ct.db, id)
	if err != nil {
		return nil, "", nil
	}
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	camp, err := camps.SelectCamp(ct.db, compta.IdCamp)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	content, err = generateExcelRecords(recs, subs)
	if err != nil {
		return nil, "", err
	}

	filename = fmt.Sprintf("Compta %s.xlsx", camp.Label())
	return content, filename, err
}

// TODO
func generateExcelRecords(recs ca.ComptaRecords, subs ca.ComptaSubrecords) ([]byte, error) {
	// panic("unimplemented")
	return nil, nil
}
