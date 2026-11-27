package espaceperso

import (
	"database/sql"
	"errors"

	filesAPI "registro/controllers/files"
	"registro/crypto"
	"registro/logic"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

var errQuotientFamilialValidated = errors.New("Cette attestation est déjà validée. Demandez au centre de la refuser avant d'en envoyer une nouvelle.")

// ensureQuotientFamilialEditable refuse toute modification d'une attestation
// déjà validée. Les attestations en attente ou refusées restent modifiables.
func ensureQuotientFamilialEditable(tx *sql.Tx, id ds.IdDossier) error {
	link, found, err := fs.SelectFileQuotientFamilialByIdDossier(tx, id)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	file, err := fs.SelectFile(tx, link.IdFile)
	if err != nil {
		return err
	}
	if file.ReviewStatus == pr.FileReviewValidated {
		return errQuotientFamilialValidated
	}
	return nil
}

// UploadQuotientFamilial téléverse l'attestation de quotient familial
// du dossier. Il n'y a au plus un fichier par dossier : un nouvel envoi
// remplace le précédent et remet son statut de validation en attente.
// Une attestation déjà validée ne peut être ni remplacée ni supprimée.
func (ct *Controller) UploadQuotientFamilial(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	header, err := c.FormFile("file")
	if err != nil {
		return err
	}
	content, name, err := filesAPI.ReadUpload(header)
	if err != nil {
		return err
	}
	out, err := ct.uploadQuotientFamilial(id, content, name)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) uploadQuotientFamilial(id ds.IdDossier, content []byte, fileName string) (logic.PublicFile, error) {
	var out logic.PublicFile
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		if err := ensureQuotientFamilialEditable(tx, id); err != nil {
			return err
		}
		link, found, err := fs.SelectFileQuotientFamilialByIdDossier(tx, id)
		if err != nil {
			return err
		}
		if found {
			if err := ct.deleteQuotientFamilialLink(tx, id, link.IdFile); err != nil {
				return err
			}
		}

		file, err := fs.File{}.Insert(tx)
		if err != nil {
			return err
		}
		if err := (fs.FileQuotientFamilial{IdFile: file.Id, IdDossier: id}).Insert(tx); err != nil {
			return err
		}
		uploaded, err := fs.UploadFile(ct.files, tx, file, content, fileName)
		if err != nil {
			return err
		}
		out = logic.NewPublicFile(ct.key, uploaded)
		return nil
	})
	return out, err
}

func (ct *Controller) DeleteQuotientFamilial(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	if err := ct.deleteQuotientFamilial(id); err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteQuotientFamilial(id ds.IdDossier) error {
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		if err := ensureQuotientFamilialEditable(tx, id); err != nil {
			return err
		}
		link, found, err := fs.SelectFileQuotientFamilialByIdDossier(tx, id)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		return ct.deleteQuotientFamilialLink(tx, id, link.IdFile)
	})
}

func (ct *Controller) deleteQuotientFamilialLink(tx *sql.Tx, id ds.IdDossier, idFile fs.IdFile) error {
	if _, err := fs.DeleteFileQuotientFamilialsByIdDossiers(tx, id); err != nil {
		return err
	}
	deleted, err := fs.DeleteFilesByIDs(tx, idFile)
	if err != nil {
		return err
	}
	return ct.files.Delete(deleted...)
}
