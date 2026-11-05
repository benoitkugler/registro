package compta

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"path/filepath"
	"sort"
	"time"

	"registro/mails"
	cps "registro/sql/camps"
	"registro/sql/compta"
	ca "registro/sql/compta"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

var errAccess = errors.New("access forbidden")

type GetRecordsOut struct {
	Equipiers map[cps.IdEquipier]string
	Records   []RecordExt
}

func (ct *Controller) GetRecords(c echo.Context) error {
	user := JWTUser(c)

	idCompta, err := utils.QueryParamInt[ca.IdComptaCamp](c, "id-compta")
	if err != nil {
		return err
	}

	if user != userAdmin && user != idCompta {
		return errAccess
	}

	out, err := ct.getRecords(idCompta)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type RecordExt struct {
	Record     ca.ComptaRecord
	Subrecords []ca.ComptaSubrecord

	Status RecordStatus
	Total  string // possibly multiples currencies
}

func newRecordExt(record ca.ComptaRecord, subrecords []ca.ComptaSubrecord) RecordExt {
	return RecordExt{record, subrecords, recordStatus(record, subrecords), recordTotal(subrecords).String()}
}

type RecordStatus uint8

const (
	Incomplete RecordStatus = iota
	Complete
	Validated
)

func recordStatus(record ca.ComptaRecord, subrecords []ca.ComptaSubrecord) RecordStatus {
	if record.IsValidated {
		return Validated
	}

	hasJustificatif, areSubrecordsValid := true, true
	for _, sub := range subrecords {
		if !sub.IdFile.Valid {
			hasJustificatif = false
		}
		if sub.Mode == ca.Cheque && sub.Serial == "" {
			areSubrecordsValid = false
		}
	}

	isComplete :=
		// justificatif is optional for other account
		(record.Account == ca.AccountAsso || hasJustificatif) && areSubrecordsValid
	if isComplete {
		return Complete
	}
	return Incomplete
}

func recordTotal(subrecords []ca.ComptaSubrecord) (out ds.MultiCurrencies) {
	for _, subrecord := range subrecords {
		out.AddM(ds.Montant{Currency: subrecord.Currency, Cent: subrecord.Montants.Total()})
	}
	return out
}

func getRecord(db ca.DB, id ca.IdComptaSubrecord) (ca.ComptaRecord, ca.ComptaSubrecord, error) {
	subrecord, err := ca.SelectComptaSubrecord(db, id)
	if err != nil {
		return ca.ComptaRecord{}, ca.ComptaSubrecord{}, utils.SQLError(err)
	}
	rec, err := ca.SelectComptaRecord(db, subrecord.IdRecord)
	if err != nil {
		return ca.ComptaRecord{}, ca.ComptaSubrecord{}, utils.SQLError(err)
	}
	return rec, subrecord, nil
}

func getRecords(db ca.DB, id ca.IdComptaCamp) (ca.ComptaRecords, ca.ComptaSubrecords, error) {
	recs, err := ca.SelectComptaRecordsByIdCamps(db, id)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	subrecords, err := ca.SelectComptaSubrecordsByIdRecords(db, recs.IDs()...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	return recs, subrecords, nil
}

func (ct *Controller) getRecords(id ca.IdComptaCamp) (GetRecordsOut, error) {
	camp, err := compta.SelectComptaCamp(ct.db, id)
	if err != nil {
		return GetRecordsOut{}, utils.SQLError(err)
	}
	recs, tmp, err := getRecords(ct.db, id)
	if err != nil {
		return GetRecordsOut{}, err
	}
	subrecords := tmp.ByIdRecord()

	out := make([]RecordExt, 0, len(recs))
	for _, rec := range recs {
		out = append(out, newRecordExt(rec, subrecords[rec.Id].Sorted()))
	}
	// newer first
	sort.Slice(out, func(i, j int) bool { return out[i].Record.Id > out[j].Record.Id })

	equipiers, personnes, _, err := cps.LoadEquipiersByCamps(ct.db, camp.IdCamp)
	if err != nil {
		return GetRecordsOut{}, err
	}
	m := make(map[cps.IdEquipier]string)
	for _, eq := range equipiers {
		m[eq.Id] = personnes[eq.IdPersonne].PrenomN()
	}

	return GetRecordsOut{m, out}, nil
}

type CreateRecordIn struct {
	IdCompta ca.IdComptaCamp
	Account  ca.AccountKind
}

// See also [Controller.UploadAndCreate]
func (ct *Controller) CreateRecord(c echo.Context) error {
	user := JWTUser(c)

	var args CreateRecordIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	if user != userAdmin && user != args.IdCompta {
		return errAccess
	}

	out, err := ct.createRecord(args)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) createRecord(args CreateRecordIn) (out RecordExt, err error) {
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		record, err := ca.ComptaRecord{IdCamp: args.IdCompta, Account: args.Account}.Insert(tx)
		if err != nil {
			return err
		}
		sub, err := ca.ComptaSubrecord{
			IdRecord: record.Id,
			Montants: ca.Montants{{Kind: ca.NoVentilation}},
			Currency: ds.Euros,
		}.Insert(tx)
		if err != nil {
			return err
		}

		out = newRecordExt(record, []ca.ComptaSubrecord{sub})

		return nil
	})

	return out, err
}

func (ct *Controller) AddSubrecord(c echo.Context) error {
	user := JWTUser(c)

	idRecord, err := utils.QueryParamInt[ca.IdComptaRecord](c, "id-record")
	if err != nil {
		return err
	}

	out, err := ct.addSubrecord(idRecord, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) addSubrecord(record ca.IdComptaRecord, user ca.IdComptaCamp) (ca.ComptaSubrecord, error) {
	group, err := ca.SelectComptaRecord(ct.db, record)
	if err != nil {
		return ca.ComptaSubrecord{}, utils.SQLError(err)
	}

	if user != userAdmin && user != group.IdCamp {
		return ca.ComptaSubrecord{}, errAccess
	}

	out, err := ca.ComptaSubrecord{
		IdRecord: record,
		Currency: ds.Euros,
		Montants: ca.Montants{{Kind: ca.NoVentilation}},
	}.Insert(ct.db)
	if err != nil {
		return ca.ComptaSubrecord{}, utils.SQLError(err)
	}
	return out, nil
}

type RecordAndSub struct {
	Record    ca.ComptaRecord
	Subrecord ca.ComptaSubrecord
}

func (ct *Controller) UpdateRecord(c echo.Context) error {
	user := JWTUser(c)

	var args RecordExt
	if err := c.Bind(&args); err != nil {
		return err
	}

	updated, err := ct.updateRecord(args, user)
	if err != nil {
		return err
	}

	return c.JSON(200, updated)
}

func (ct *Controller) updateRecord(in RecordExt, user ca.IdComptaCamp) (out RecordExt, err error) {
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		currentGroup, err := ca.SelectComptaRecord(tx, in.Record.Id)
		if err != nil {
			return utils.SQLError(err)
		}
		currents, err := ca.SelectComptaSubrecordsByIdRecords(tx, currentGroup.Id)
		if err != nil {
			return utils.SQLError(err)
		}

		if user != userAdmin && user != currentGroup.IdCamp {
			return errAccess
		}

		// remove validation when a "sensitive" field is changed by a director
		hasChanged := in.Record.IsReceipe != currentGroup.IsReceipe
		for _, change := range in.Subrecords {
			if err = checkVentilations(change.Montants, in.Record.IsReceipe); err != nil {
				return err
			}

			current := currents[change.Id]

			hasChanged = hasChanged || (change.Mode != current.Mode || change.Serial != current.Serial || !change.Montants.Equals(current.Montants))

			current.Date = change.Date
			current.Description = change.Description
			current.Mode = change.Mode
			current.Serial = change.Serial
			current.Currency = change.Currency
			current.Montants = change.Montants

			current, err = current.Update(tx)
			if err != nil {
				return utils.SQLError(err)
			}

			currents[current.Id] = current
		}

		// remove validation when a "sensitive" field is changed by a director
		if user != userAdmin && hasChanged {
			in.Record.IsValidated = false
		}

		currentGroup.IsValidated = in.Record.IsValidated
		currentGroup.Account = in.Record.Account
		currentGroup.IsReceipe = in.Record.IsReceipe
		currentGroup.SpentBy = in.Record.SpentBy

		currentGroup, err = currentGroup.Update(tx)
		if err != nil {
			return utils.SQLError(err)
		}

		out = newRecordExt(currentGroup, currents.Sorted())

		return nil
	})

	return out, err
}

type MergeRecordToIn struct {
	Source ca.IdComptaRecord
	Target ca.IdComptaRecord
}

// MergeRecordTo adds the subrecords of [Source] to [Target],
// and deletes [Source].
// The validation of [Target] is removed if the operation is triggered
// by a director.
func (ct *Controller) MergeRecordTo(c echo.Context) error {
	user := JWTUser(c)

	var args MergeRecordToIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.mergeRecordTo(args, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) mergeRecordTo(args MergeRecordToIn, user ca.IdComptaCamp) (out RecordExt, err error) {
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		target, err := ca.SelectComptaRecord(tx, args.Target)
		if err != nil {
			return utils.SQLError(err)
		}
		source, err := ca.SelectComptaRecord(tx, args.Source)
		if err != nil {
			return utils.SQLError(err)
		}

		if user != userAdmin && (user != target.IdCamp || user != source.IdCamp) {
			return errAccess
		}

		// "move" the subrecords
		toMove, err := ca.SelectComptaSubrecordsByIdRecords(tx, source.Id)
		if err != nil {
			return utils.SQLError(err)
		}
		for _, sub := range toMove {
			sub.IdRecord = target.Id
			_, err = sub.Update(tx)
			if err != nil {
				return utils.SQLError(err)
			}
		}

		// remove the source
		_, err = ca.DeleteComptaRecordById(tx, source.Id)
		if err != nil {
			return utils.SQLError(err)
		}

		// remove the validation
		if user != userAdmin {
			target.IsValidated = false
			target, err = target.Update(tx)
			if err != nil {
				return utils.SQLError(err)
			}
		}

		// load all subrecords
		m, err := ca.SelectComptaSubrecordsByIdRecords(tx, target.Id)
		if err != nil {
			return utils.SQLError(err)
		}

		out = newRecordExt(target, m.Sorted())

		return nil
	})
	return
}

type AddRecordReviewIn struct {
	IdRecord ca.IdComptaRecord
	Text     string
}

func (ct *Controller) AddRecordReview(c echo.Context) error {
	user := JWTUser(c)

	var args AddRecordReviewIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.addRecordReview(args, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) addRecordReview(args AddRecordReviewIn, user ca.IdComptaCamp) (ca.ComptaRecord, error) {
	record, err := ca.SelectComptaRecord(ct.db, args.IdRecord)
	if err != nil {
		return ca.ComptaRecord{}, utils.SQLError(err)
	}
	compta, err := ca.SelectComptaCamp(ct.db, record.IdCamp)
	if err != nil {
		return ca.ComptaRecord{}, utils.SQLError(err)
	}
	camp, directorM, hasDirector, err := director(ct.db, compta)
	if err != nil {
		return ca.ComptaRecord{}, utils.SQLError(err)
	}

	if user != userAdmin && user != record.IdCamp {
		return ca.ComptaRecord{}, errAccess
	}

	record.Reviews = append(record.Reviews, ca.Review{Time: time.Now(), Message: args.Text, IsAdmin: user == userAdmin})
	record, err = record.Update(ct.db)
	if err != nil {
		return ca.ComptaRecord{}, utils.SQLError(err)
	}

	// avoid sending to many emails
	const delta = time.Hour
	lastSentPtr := &compta.LastSentMailDirecteur
	if user == userAdmin {
		lastSentPtr = &compta.LastSentMailAdmin
	}
	if (*lastSentPtr).Add(delta).Before(time.Now()) {
		*lastSentPtr = time.Now()
		_, err = compta.Update(ct.db)
		if err != nil {
			return ca.ComptaRecord{}, utils.SQLError(err)
		}

		message := fmt.Sprintf(`Bonjour, <br/><br/>
			
		Vous avez reçu une nouvelle notification concernant la compta du séjour <i>%s</i>. <br/><br/>

		Merci et bon courage !
		`, camp.Label())
		// notify the receipient
		if user == userAdmin && hasDirector {
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(directorM, "Comptabilité des séjours", message, nil, nil)
			if err != nil {
				return record, fmt.Errorf("Envoi de la notification: %w", err)
			}
		} else if user != userAdmin {
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(ct.adminNotificationMail, "Comptabilité des séjours", message, nil, nil)
			if err != nil {
				return record, fmt.Errorf("Envoi de la notification: %w", err)
			}
		}
	}

	return record, nil
}

func (ct *Controller) DeleteSubrecord(c echo.Context) error {
	user := JWTUser(c)

	idR, err := utils.QueryParamInt[ca.IdComptaSubrecord](c, "id")
	if err != nil {
		return err
	}

	err = ct.deleteSubrecord(idR, user)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) deleteSubrecord(id ca.IdComptaSubrecord, user ca.IdComptaCamp) error {
	rec, subrecord, err := getRecord(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	compta, err := ca.SelectComptaCamp(ct.db, rec.IdCamp)
	if err != nil {
		return utils.SQLError(err)
	}

	if user != userAdmin && user != rec.IdCamp {
		return errAccess
	}

	_, err = ca.DeleteComptaSubrecordById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}

	// remove empty groups
	others, err := ca.SelectComptaSubrecordsByIdRecords(ct.db, rec.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if len(others) == 0 {
		_, err = ca.DeleteComptaRecordById(ct.db, rec.Id)
		if err != nil {
			return utils.SQLError(err)
		}
	}

	if subrecord.IdFile.Valid {
		// remove associated files
		go func() {
			err := ct.files.remove(ct.db, compta.FileStatus, subrecord.IdFile.Id)
			if err != nil {
				log.Println("deleteSubrecord:", err)
			}
		}()
	}

	return nil
}

// UploadAndCreate creates a new Record with the given file.
func (ct *Controller) UploadAndCreate(c echo.Context) error {
	user := JWTUser(c)

	idCompta, err := utils.FormValueInt[ca.IdComptaCamp](c, "idCompta")
	if err != nil {
		return fmt.Errorf("invalid ID parameter: %s", err)
	}

	header, err := c.FormFile("file")
	if err != nil {
		return err
	}

	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	rec, err := ct.uploadAndCreate(idCompta, header.Filename, file, user)
	if err != nil {
		return err
	}

	return c.JSON(200, rec)
}

func (ct *Controller) uploadAndCreate(idCompta ca.IdComptaCamp, fileName string, fileContent io.Reader, user ca.IdComptaCamp) (out RecordExt, err error) {
	if user != userAdmin && user != idCompta {
		return out, errAccess
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		compta, err := ca.SelectComptaCamp(tx, idCompta)
		if err != nil {
			return err
		}

		// create the file item
		meta, err := ct.files.createFile(tx, compta.FileStatus, fileName, fileContent)
		if err != nil {
			return err
		}

		// create the record
		record, err := ca.ComptaRecord{IdCamp: idCompta}.Insert(tx)
		if err != nil {
			return utils.SQLError(err)
		}
		sub, err := ca.ComptaSubrecord{
			IdRecord:    record.Id,
			Montants:    ca.Montants{{Kind: ca.NoVentilation}},
			Description: "A compléter",
			Currency:    ds.Euros,
			IdFile:      ca.NewOptIdComptaFile(meta.Id),
		}.Insert(tx)
		if err != nil {
			return utils.SQLError(err)
		}

		out = newRecordExt(record, []ca.ComptaSubrecord{sub})

		return nil
	})
	return out, err
}

func (ct *Controller) UploadJustificatif(c echo.Context) error {
	user := JWTUser(c)

	idRecord, err := utils.FormValueInt[ca.IdComptaSubrecord](c, "idSubrecord")
	if err != nil {
		return fmt.Errorf("invalid ID parameter: %s", err)
	}

	header, err := c.FormFile("file")
	if err != nil {
		return err
	}

	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	rec, err := ct.uploadJustificatif(idRecord, header.Filename, file, user)
	if err != nil {
		return err
	}

	return c.JSON(200, rec)
}

func (ct *Controller) uploadJustificatif(id ca.IdComptaSubrecord, filename string, file io.Reader, user ca.IdComptaCamp) (RecordAndSub, error) {
	group, rec, err := getRecord(ct.db, id)
	if err != nil {
		return RecordAndSub{}, fmt.Errorf("Erreur interne: %w", err)
	}
	compta, err := ca.SelectComptaCamp(ct.db, group.IdCamp)
	if err != nil {
		return RecordAndSub{}, utils.SQLError(err)
	}

	if user != userAdmin && user != group.IdCamp {
		return RecordAndSub{}, errAccess
	}

	if rec.IdFile.Valid {
		// remove current files
		go ct.files.remove(ct.db, compta.FileStatus, rec.IdFile.Id)
	}

	meta, err := ct.files.createFile(ct.db, compta.FileStatus, filename, file)
	if err != nil {
		return RecordAndSub{}, err
	}

	if user != userAdmin {
		group.IsValidated = false // remove the validation since the justificatif has changed
		group, err = group.Update(ct.db)
		if err != nil {
			return RecordAndSub{}, utils.SQLError(err)
		}
	}

	rec.IdFile = ca.NewOptIdComptaFile(meta.Id)
	rec, err = rec.Update(ct.db)
	if err != nil {
		return RecordAndSub{}, utils.SQLError(err)
	}

	return RecordAndSub{group, rec}, nil
}

func (ct *Controller) RemoveJustificatif(c echo.Context) error {
	user := JWTUser(c)

	idR, err := utils.QueryParamInt[ca.IdComptaSubrecord](c, "id-subrecord")
	if err != nil {
		return err
	}

	rec, err := ct.removeJustificatif(idR, user)
	if err != nil {
		return err
	}

	return c.JSON(200, rec)
}

func (ct *Controller) removeJustificatif(id ca.IdComptaSubrecord, user ca.IdComptaCamp) (RecordAndSub, error) {
	group, rec, err := getRecord(ct.db, id)
	if err != nil {
		return RecordAndSub{}, fmt.Errorf("Erreur interne: %w", err)
	}
	compta, err := ca.SelectComptaCamp(ct.db, group.IdCamp)
	if err != nil {
		return RecordAndSub{}, utils.SQLError(err)
	}

	if user != userAdmin && user != group.IdCamp {
		return RecordAndSub{}, errAccess
	}

	err = ct.files.remove(ct.db, compta.FileStatus, rec.IdFile.Id)
	if err != nil {
		return RecordAndSub{}, fmt.Errorf("Erreur interne: %w", err)
	}

	if user != userAdmin {
		group.IsValidated = false // remove the validation since the justificatif is no more
		group, err = group.Update(ct.db)
		if err != nil {
			return RecordAndSub{}, utils.SQLError(err)
		}
	}

	rec.IdFile = ca.OptIdComptaFile{}
	rec, err = rec.Update(ct.db)
	if err != nil {
		return RecordAndSub{}, fmt.Errorf("Erreur interne: %w", err)
	}

	return RecordAndSub{group, rec}, nil
}

func (ct *Controller) DownloadJustificatif(c echo.Context) error {
	user := JWTUser(c)

	idR, err := utils.QueryParamInt[ca.IdComptaSubrecord](c, "id-subrecord")
	if err != nil {
		return err
	}

	fileBytes, filename, err := ct.getJustificatif(idR, user, false)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, utils.AttachementHeader(filename))
	return c.Blob(200, mime.TypeByExtension(filepath.Ext(filename)), fileBytes)
}

func (ct *Controller) GetMiniature(c echo.Context) error {
	const pathMiniatureError = "static/images/default_miniature.png"

	user := JWTUser(c)

	idR, err := utils.QueryParamInt[ca.IdComptaSubrecord](c, "id-subrecord")
	if err != nil {
		return err
	}

	min, _, err := ct.getJustificatif(idR, user, true)
	if err != nil {
		return c.File(pathMiniatureError)
	}

	return c.Blob(200, mime.TypeByExtension(".png"), min)
}

func (ct *Controller) getJustificatif(id ca.IdComptaSubrecord, user ca.IdComptaCamp, isMiniature bool) ([]byte, string, error) {
	group, rec, err := getRecord(ct.db, id)
	if err != nil {
		return nil, "", fmt.Errorf("Erreur interne: %w", err)
	}
	compta, err := ca.SelectComptaCamp(ct.db, group.IdCamp)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	if user != userAdmin && user != group.IdCamp {
		return nil, "", errAccess
	}

	return ct.files.download(ct.db, compta.FileStatus, rec.IdFile.Id, isMiniature)
}
