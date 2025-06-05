package espaceperso

import (
	"errors"
	"fmt"

	filesAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"

	"github.com/labstack/echo/v4"
)

type Documents struct {
	Files []FilesCamp
}

type FilesCamp struct {
	Camp      string
	Generated []GeneratedFile
	Files     []filesAPI.PublicFile
}

type GeneratedFile struct {
	NomClient string
	Key       string // crypted camp ID and Kind
}

func (ct *Controller) GetDocuments(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	out, err := ct.loadDocuments(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadDocuments(id ds.IdDossier) (Documents, error) {
	dossier, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return Documents{}, err
	}
	camps := dossier.CampsInscrits()
	for _, camp := range camps {
		item := FilesCamp{Camp: camp.Label()}
		// generated files
		if camp.DocumentsToShow.ListeVetements {
			key, err := ct.key.EncryptJSON(generatedFileKey{IdCamp: camp.Id, Kind: listeVetements})
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, GeneratedFile{
				NomClient: fmt.Sprintf("Liste de vêtements %s.pdf", camp.Label()),
				Key:       key,
			})
		}
		if camp.DocumentsToShow.ListeParticipants {
			key, err := ct.key.EncryptJSON(generatedFileKey{IdCamp: camp.Id, Kind: listeParticipants})
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, GeneratedFile{
				NomClient: fmt.Sprintf("Liste des participants %s.pdf", camp.Label()),
				Key:       key,
			})
		}

	}

	return Documents{}, nil
}

type generatedKind uint8

const (
	_ generatedKind = iota
	listeVetements
	listeParticipants
)

type generatedFileKey struct {
	IdCamp cps.IdCamp
	Kind   generatedKind
}
