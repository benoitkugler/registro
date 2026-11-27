package backoffice

import (
	"database/sql"
	"errors"
	"time"

	"registro/caf"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// QuotientFamilialDetectionStatus décrit l'issue de la détection
// automatique du quotient familial dans l'attestation téléversée.
type QuotientFamilialDetectionStatus uint8

const (
	QFDetectionNoFile      QuotientFamilialDetectionStatus = iota // Aucune attestation
	QFDetectionSuccess                                            // QF détecté
	QFDetectionUnreadable                                         // Fichier illisible
	QFDetectionInvalidPDF                                         // PDF invalide
	QFDetectionNoText                                             // PDF sans texte
	QFDetectionUnsupported                                        // Attestation non reconnue
	QFDetectionAmbiguous                                          // Résultat ambigu
)

type QuotientFamilialDetection struct {
	Status           QuotientFamilialDetectionStatus
	Message          string // détail en cas d'échec
	QuotientFamilial int
	Mois             time.Month
	Annee            int
}

// detectQuotientFamilial analyse l'attestation avec le parseur CAF.
// Un échec de détection n'empêche pas le chargement du dossier : le
// backoffice affiche le statut et laisse l'administrateur saisir le QF.
func detectQuotientFamilial(dossier logic.DossierExt, files fs.FileSystem) QuotientFamilialDetection {
	if !dossier.HasQuotientFamilialFile {
		return QuotientFamilialDetection{
			Status:  QFDetectionNoFile,
			Message: "Aucune attestation téléversée.",
		}
	}
	content, err := files.Load(dossier.QuotientFamilialFile.Id, false)
	if err != nil {
		return QuotientFamilialDetection{
			Status:  QFDetectionUnreadable,
			Message: err.Error(),
		}
	}
	att, err := caf.ParseAttestation(content)
	if err != nil {
		return QuotientFamilialDetection{Status: detectionStatus(err), Message: err.Error()}
	}
	return QuotientFamilialDetection{
		Status:           QFDetectionSuccess,
		QuotientFamilial: att.QuotientFamilial,
		Mois:             att.Mois,
		Annee:            att.Annee,
	}
}

func detectionStatus(err error) QuotientFamilialDetectionStatus {
	switch {
	case errors.Is(err, caf.ErrInvalidPDF):
		return QFDetectionInvalidPDF
	case errors.Is(err, caf.ErrNoExtractableText):
		return QFDetectionNoText
	case errors.Is(err, caf.ErrAmbiguous):
		return QFDetectionAmbiguous
	case errors.Is(err, caf.ErrUnsupportedDocument):
		return QFDetectionUnsupported
	default:
		return QFDetectionInvalidPDF
	}
}

type QuotientFamilialValidateIn struct {
	IdDossier        ds.IdDossier
	QuotientFamilial int // valeur retenue, éventuellement corrigée par l'administrateur
	Accepted         bool
}

// DossiersValidateQuotientFamilial valide ou refuse l'attestation,
// et applique le quotient familial retenu aux participants du dossier.
func (ct *Controller) DossiersValidateQuotientFamilial(c echo.Context) error {
	var args QuotientFamilialValidateIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	if err := ct.validateQuotientFamilial(args); err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) validateQuotientFamilial(args QuotientFamilialValidateIn) error {
	if args.QuotientFamilial < 0 {
		return errors.New("quotient familial invalide")
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		link, found, err := fs.SelectFileQuotientFamilialByIdDossier(tx, args.IdDossier)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("aucune attestation de quotient familial pour ce dossier")
		}
		file, err := fs.SelectFile(tx, link.IdFile)
		if err != nil {
			return err
		}
		if args.Accepted {
			file.ReviewStatus = pr.FileReviewValidated
		} else {
			file.ReviewStatus = pr.FileReviewRejected
		}
		if _, err := file.Update(tx); err != nil {
			return err
		}
		if !args.Accepted {
			return nil
		}

		participants, err := cps.SelectParticipantsByIdDossiers(tx, args.IdDossier)
		if err != nil {
			return err
		}
		for _, participant := range participants {
			participant.QuotientFamilial = args.QuotientFamilial
			if _, err := participant.Update(tx); err != nil {
				return err
			}
		}
		return nil
	})
}
