package espaceperso

import (
	"errors"
	"fmt"

	filesAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Documents struct {
	FilesToRead   []FilesCamp
	FilesToUpload []DemandesPersonne
}

type FilesCamp struct {
	Camp      string
	Generated []GeneratedFile
	Files     []filesAPI.PublicFile
}

type DemandesPersonne struct {
	Personne string
	Demandes []DemandePersonne
}

type DemandePersonne struct {
	Demande     fs.Demande
	DemandeFile filesAPI.PublicFile
	Uploaded    []filesAPI.PublicFile
}

type GeneratedFile struct {
	NomClient string
	Key       string // crypted camp ID and Kind
}

func (ct *Controller) LoadDocuments(c echo.Context) error {
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
	links, err := fs.SelectFileCampsByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}
	byCamp := links.ByIdCamp()
	campFiles, err := fs.SelectFiles(ct.db, links.IdFiles()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}

	var out Documents
	for _, camp := range camps {
		item := FilesCamp{Camp: camp.Label()}
		// generated files
		if camp.DocumentsToShow.ListeVetements {
			key, err := filesAPI.CampDocumentKey(ct.key, camp.Id, filesAPI.ListeVetements)
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, GeneratedFile{
				NomClient: fmt.Sprintf("Liste de vêtements %s.pdf", camp.Label()),
				Key:       key,
			})
		}
		if camp.DocumentsToShow.ListeParticipants {
			key, err := filesAPI.CampDocumentKey(ct.key, camp.Id, filesAPI.ListeParticipants)
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, GeneratedFile{
				NomClient: fmt.Sprintf("Liste des participants %s.pdf", camp.Label()),
				Key:       key,
			})
		}
		// other
		for _, link := range byCamp[camp.Id] {
			if (link.IsLettre && camp.DocumentsToShow.LettreDirecteur) ||
				!link.IsLettre {
				item.Files = append(item.Files, filesAPI.NewPublicFile(ct.key, campFiles[link.IdFile]))
			}
		}

		out.FilesToRead = append(out.FilesToRead, item)
	}

	links2, err := fs.SelectDemandeCampsByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}
	demandesByCamp := links2.ByIdCamp()
	personnesFiles, demandes, err := filesAPI.LoadFilesPersonnes(ct.db, ct.key, links2.IdDemandes(), dossier.Participants.IdPersonnes()...)
	if err != nil {
		return Documents{}, err
	}
	demandesFiles, err := fs.SelectFiles(ct.db, demandes.IdFiles()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}

	byPersonne := dossier.Participants.ByIdPersonne()
	for _, personne := range dossier.Personnes()[1:] {
		item := DemandesPersonne{Personne: personne.PrenomN()}
		for _, part := range byPersonne[personne.Id] {
			if part.Statut != cps.Inscrit {
				continue
			}
			// add demandes from the camp
			for _, link := range demandesByCamp[part.IdCamp] {
				demande := demandes[link.IdDemande]
				demandeF := DemandePersonne{
					Demande:  demande,
					Uploaded: personnesFiles[demande.Id][personne.Id],
				}
				if fi := demande.IdFile; fi.Valid {
					demandeF.DemandeFile = filesAPI.NewPublicFile(ct.key, demandesFiles[fi.Id])
				}
				item.Demandes = append(item.Demandes, demandeF)
			}
		}
		// do not include empty lists
		if len(item.Demandes) != 0 {
			out.FilesToUpload = append(out.FilesToUpload, item)
		}
	}

	return out, nil
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
