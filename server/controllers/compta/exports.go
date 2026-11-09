package compta

import (
	"fmt"
	"iter"
	"path/filepath"

	"registro/controllers/files"
	cps "registro/sql/camps"
	ca "registro/sql/compta"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// TODO: check the code, add tests
// Export excels

// DownloadEcritures renvoie un fichier Excel contenant les écritures comptables
func (ct *Controller) DownloadEcritures(c echo.Context) error {
	user := JWTUser(c)

	idCompta, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != idCompta {
		return errAccess
	}

	fileBytes, filename, err := ct.getEcritures(idCompta)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, utils.AttachementHeader(filename))
	return c.Blob(200, "application/vnd.ms-excel", fileBytes)
}

func (ct *Controller) getEcritures(idCompta ca.IdComptaCamp) ([]byte, string, error) {
	compta, err := ca.SelectComptaCamp(ct.db, idCompta)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	ld, err := newComptaLoader(ct.db, compta)
	if err != nil {
		return nil, "", err
	}

	ecr := ld.ecritures()
	buf, err := generateExcelEcritures(ecr)
	if err != nil {
		return nil, "", fmt.Errorf("Erreur pendant la création du fichier Excel: %w", err)
	}

	filename := fmt.Sprintf("Ecritures %s.xlsx", ld.camp.Label())
	return buf.Bytes(), filename, nil
}

type BilanCampsIn struct {
	Ids []ca.IdComptaCamp
}

// DownloadBilanCamps renvoie un fichier Excel contenant les détails
// des camps demandés
func (ct *Controller) DownloadBilanCamps(c echo.Context) error {
	user := JWTUser(c)

	if user != userAdmin {
		return errAccess
	}

	var args BilanCampsIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	fileBytes, filename, err := ct.getBilanCamps(args)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, utils.AttachementHeader(filename))
	return c.Blob(200, "application/vnd.ms-excel", fileBytes)
}

func (ct *Controller) getBilanCamps(args BilanCampsIn) ([]byte, string, error) {
	loaders, err := newComptaLoaders(ct.db, args.Ids)
	if err != nil {
		return nil, "", err
	}

	bc := make(bilanCamp, len(loaders))
	for i, ld := range loaders {
		bc[i] = ld.compile()
	}

	buf, err := generateExcelCamps(bc)
	if err != nil {
		return nil, "", fmt.Errorf("Erreur pendant la création du fichier Excel: %w", err)
	}

	return buf.Bytes(), "Bilan des séjours.xlsx", nil
}

// DownloadRecords renvoie un fichier Excel des dépenses.
func (ct *Controller) DownloadRecords(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != id {
		return errAccess
	}

	fileBytes, filename, err := ct.getRecordsArchive(id)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, utils.AttachementHeader(filename))
	return c.Blob(200, "application/vnd.ms-excel", fileBytes)
}

// DownloadRecordsAndFilesStream renvoie une archive zip contenant un fichier Excel
// des dépenses et les justificatifs associés, par morceaux.
func (ct *Controller) DownloadRecordsAndFilesStream(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != id {
		return errAccess
	}

	name, iter, err := ct.streamRecordsArchive(id)
	if err != nil {
		return err
	}
	err = files.StreamZip(c.Response(), name, iter)
	return err
}

func (ct *Controller) streamRecordsArchive(id ca.IdComptaCamp) (string, iter.Seq2[files.ZipItem, error], error) {
	groups, recs, err := getRecords(ct.db, id)
	if err != nil {
		return "", nil, err
	}
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return "", nil, utils.SQLError(err)
	}
	camp, err := cps.SelectCamp(ct.db, compta.IdCamp)
	if err != nil {
		return "", nil, utils.SQLError(err)
	}

	excel, err := generateExcelRecords(groups, recs)
	if err != nil {
		return "", nil, err
	}

	filename := fmt.Sprintf("Compta %s.zip", camp.Label())
	return filename, func(yield func(files.ZipItem, error) bool) {
		if !yield(files.ZipItem{Name: "compta.xlsx", Content: excel.Bytes()}, nil) {
			return
		}
		for _, rec := range recs {
			if !rec.IdFile.Valid {
				continue
			}
			content, originalName, err := ct.files.download(ct.db, compta.FileStatus, rec.IdFile.Id, false)
			if err != nil {
				yield(files.ZipItem{}, err)
				return
			}
			item := files.ZipItem{
				Name:    fmt.Sprintf("justificatif_%d%s", rec.Id, filepath.Ext(originalName)),
				Content: content,
			}
			if !yield(item, nil) {
				return
			}
		}
	}, nil
}

func (ct *Controller) getRecordsArchive(id ca.IdComptaCamp) (content []byte, filename string, err error) {
	recs, subs, err := getRecords(ct.db, id)
	if err != nil {
		return nil, "", err
	}
	compta, err := ca.SelectComptaCamp(ct.db, id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	camp, err := cps.SelectCamp(ct.db, compta.IdCamp)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	excel, err := generateExcelRecords(recs, subs)
	if err != nil {
		return nil, "", fmt.Errorf("Erreur pendant la création du fichier Excel: %w", err)
	}

	content, filename = excel.Bytes(), fmt.Sprintf("Compta %s.xlsx", camp.Label())
	return content, filename, err
}
