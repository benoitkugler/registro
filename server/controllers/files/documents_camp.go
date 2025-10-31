package files

import (
	"database/sql"
	"fmt"

	"registro/config"
	"registro/crypto"
	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	"registro/sql/files"
	"registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type GenDocumentKind uint8

const (
	_ GenDocumentKind = iota
	ListeVetements
	ListeParticipants
)

type generatedFileKey struct {
	IdCamp cps.IdCamp
	Kind   GenDocumentKind
}

type GeneratedFile struct {
	NomClient string
	Key       string // crypted camp ID and Kind
}

// CampDocument returns a key identifying generated documents
func CampDocument(key crypto.Encrypter, camp cps.Camp, kind GenDocumentKind) (out GeneratedFile, err error) {
	out.Key, err = key.EncryptJSON(generatedFileKey{camp.Id, kind})
	if err != nil {
		return GeneratedFile{}, err
	}
	switch kind {
	case ListeVetements:
		out.NomClient = fmt.Sprintf("Liste de vêtements %s.pdf", camp.Label())
	case ListeParticipants:
		out.NomClient = fmt.Sprintf("Liste des participants %s.pdf", camp.Label())
	}
	return out, nil
}

func (ct *Controller) RenderDocumentCamp(c echo.Context) error {
	documentToken := c.QueryParam("documentToken")
	isMiniature := utils.QueryParamBool(c, "isMiniature")
	content, name, err := ct.renderDocument(documentToken, isMiniature)
	if err != nil {
		return err
	}
	mimeType := SetBlobHeader(c, content, name)
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) renderDocument(documentToken string, isMiniature bool) (content []byte, name string, err error) {
	var args generatedFileKey
	if err = ct.key.DecryptJSON(documentToken, &args); err != nil {
		return nil, "", err
	}
	switch args.Kind {
	case ListeVetements:
		content, name, err = renderListeVetements(ct.db, ct.asso, args.IdCamp)
	case ListeParticipants:
		content, name, err = renderListeParticipants(ct.db, ct.asso, args.IdCamp)
	}
	if err != nil {
		return nil, "", err
	}
	if isMiniature {
		content, err = files.ComputeMiniaturePDF(content)
		if err != nil {
			return nil, "", err
		}
		name = "miniature.png"
	}

	return
}

func renderListeVetements(db *sql.DB, asso config.Asso, id cps.IdCamp) ([]byte, string, error) {
	camp, err := cps.SelectCamp(db, id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	content, err := pdfcreator.CreateListeVetements(asso, camp.Vetements, camp.Label())
	if err != nil {
		return nil, "", err
	}
	return content, fmt.Sprintf("Liste de vêtements - %s.pdf", camp.Label()), nil
}

func renderListeParticipants(db *sql.DB, asso config.Asso, id cps.IdCamp) ([]byte, string, error) {
	camp, err := cps.LoadCamp(db, id)
	if err != nil {
		return nil, "", err
	}
	dossiers, err := dossiers.SelectDossiers(db, camp.IdDossiers()...)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	responsables, err := personnes.SelectPersonnes(db, dossiers.IdResponsables()...)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	var participants []pdfcreator.Participant
	for _, part := range camp.Participants(true) {
		dossier := dossiers[part.Participant.IdDossier]
		respo := responsables[dossier.IdResponsable]
		mail := part.Personne.Mail
		if mail == "" {
			mail = respo.Mail
		}
		commune := part.Personne.Ville
		if commune == "" {
			commune = respo.Ville
		}
		participants = append(participants, pdfcreator.Participant{
			NomPrenom:   part.Personne.NOMPrenom(),
			Responsable: respo.NOMPrenom(),
			Mail:        mail, Commune: commune,
		})
	}
	content, err := pdfcreator.CreateListeParticipants(asso, participants, camp.Camp.Label())
	if err != nil {
		return nil, "", err
	}
	return content, fmt.Sprintf("Liste des participants - %s.pdf", camp.Camp.Label()), nil
}
