package espaceperso

import (
	"errors"
	"slices"

	filesAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Documents struct {
	FilesToRead   []FilesCamp
	FilesToUpload []DemandesPersonne
	ToFillCount   int
}

func (docs *Documents) setToFillCount() {
	toFill := 0
	for _, personne := range docs.FilesToUpload {
		for _, demande := range personne.Demandes {
			if len(demande.Uploaded) == 0 {
				toFill++
			}
		}
	}
	docs.ToFillCount = toFill
}

type FilesCamp struct {
	Camp      string
	Generated []filesAPI.GeneratedFile
	Files     []filesAPI.PublicFile
}

type DemandesPersonne struct {
	IdPersonne pr.IdPersonne
	Personne   string
	Demandes   []DemandePersonne
}

type DemandePersonne struct {
	Demande     fs.Demande
	DemandeFile filesAPI.PublicFile
	Uploaded    []filesAPI.PublicFile
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
	return loadDocuments(ct.db, ct.key, dossier)
}

func loadDocuments(db ds.DB, key crypto.Encrypter, dossier logic.Dossier) (Documents, error) {
	camps := dossier.CampsInscrits()
	links, err := fs.SelectFileCampsByIdCamps(db, camps.IDs()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}
	byCamp := links.ByIdCamp()
	campFiles, err := fs.SelectFiles(db, links.IdFiles()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}

	var out Documents
	for _, camp := range camps {
		item := FilesCamp{Camp: camp.Label()}
		// other files
		for _, link := range byCamp[camp.Id] {
			if (link.IsLettre && camp.DocumentsToShow.LettreDirecteur) ||
				!link.IsLettre {
				item.Files = append(item.Files, filesAPI.NewPublicFile(key, campFiles[link.IdFile]))
			}
		}

		// generated files
		if camp.DocumentsToShow.ListeVetements {
			doc, err := filesAPI.CampDocument(key, camp, filesAPI.ListeVetements)
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, doc)
		}
		if camp.DocumentsToShow.ListeParticipants {
			doc, err := filesAPI.CampDocument(key, camp, filesAPI.ListeParticipants)
			if err != nil {
				return Documents{}, err
			}
			item.Generated = append(item.Generated, doc)
		}

		// do not include empty lists
		if len(item.Generated)+len(item.Files) != 0 {
			out.FilesToRead = append(out.FilesToRead, item)
		}
	}

	links2, err := fs.SelectDemandeCampsByIdCamps(db, camps.IDs()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}
	demandesByCamp := links2.ByIdCamp()
	personnesFiles, demandes, err := filesAPI.LoadFilesPersonnes(db, key, links2.IdDemandes(), dossier.Participants.IdPersonnes()...)
	if err != nil {
		return Documents{}, err
	}
	demandesFiles, err := fs.SelectFiles(db, demandes.IdFiles()...)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}

	byPersonne := dossier.Participants.ByIdPersonne()
	for _, personne := range dossier.Personnes()[1:] {
		item := DemandesPersonne{IdPersonne: personne.Id, Personne: personne.PrenomN()}
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
					demandeF.DemandeFile = filesAPI.NewPublicFile(key, demandesFiles[fi.Id])
				}
				item.Demandes = append(item.Demandes, demandeF)
			}
		}
		// do not include empty lists
		if len(item.Demandes) != 0 {
			out.FilesToUpload = append(out.FilesToUpload, item)
		}
	}

	out.setToFillCount()
	return out, nil
}

func (ct *Controller) UploadDocument(c echo.Context) error {
	token := c.QueryParam("token")
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	idPersonne, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	header, err := c.FormFile("file")
	if err != nil {
		return err
	}
	content, filename, err := filesAPI.ReadUpload(header)
	if err != nil {
		return err
	}
	out, err := ct.uploadDocument(idDossier, idDemande, idPersonne, content, filename)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) uploadDocument(idDossier ds.IdDossier, idDemande fs.IdDemande, idPersonne pr.IdPersonne,
	content []byte, filename string,
) (filesAPI.PublicFile, error) {
	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return filesAPI.PublicFile{}, err
	}
	// basic security check
	if hasPersonne := slices.Contains(dossier.Participants.IdPersonnes(), idPersonne); !hasPersonne {
		return filesAPI.PublicFile{}, errors.New("access forbidden")
	}
	file, err := filesAPI.SaveFileFor(ct.files, ct.db, idPersonne, idDemande, content, filename)
	if err != nil {
		return filesAPI.PublicFile{}, err
	}
	return filesAPI.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DeleteDocument(c echo.Context) error {
	key := c.QueryParam("key")
	err := filesAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}
