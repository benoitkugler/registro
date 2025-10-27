package espaceperso

import (
	"errors"
	"slices"
	"time"

	filesAPI "registro/controllers/files"
	"registro/crypto"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Charte struct {
	Id       pr.IdPersonne
	Personne string
	Accepted bool
}

type Documents struct {
	FilesToRead   []FilesCamp
	FilesToUpload []DemandesPersonne // including vaccins
	Fiches        []FichesanitaireExt
	Chartes       []Charte

	NewCount int
}

func (docs *Documents) setCounts(lastLoaded time.Time, events logic.Events) {
	toFill := 0
	for _, personne := range docs.FilesToUpload {
		for _, demande := range personne.Demandes {
			if len(demande.Uploaded) == 0 {
				toFill++
			}
		}
	}

	// heuristic to count the documents to read
	campsToRead := make(utils.Set[cps.IdCamp])
	for content := range logic.IterEventsBy[logic.CampDocs](events) {
		sendAt := content.Event.Created
		// check if the docs have been sent AFTER the last time the parent
		// was here
		if sendAt.After(lastLoaded) {
			campsToRead.Add(content.Content.IdCamp)
		}
	}
	// sum, restricted to the camps
	toRead := 0
	for _, camp := range docs.FilesToRead {
		if campsToRead.Has(camp.idCamp) {
			toRead += len(camp.Files) + len(camp.Generated)
		}
	}

	// chartes
	chartesCount := 0
	for _, charte := range docs.Chartes {
		if !charte.Accepted {
			chartesCount++
		}
	}

	// fiches sanitaires
	fichesCount := 0
	for _, fiche := range docs.Fiches {
		if fiche.IsLocked || fiche.State != pr.UpToDate {
			fichesCount++
		}
	}

	docs.NewCount = toRead + toFill + fichesCount + chartesCount
}

type FilesCamp struct {
	idCamp    cps.IdCamp
	Camp      string
	Generated []filesAPI.GeneratedFile
	Files     []logic.PublicFile
}

type DemandesPersonne struct {
	IdPersonne pr.IdPersonne
	Personne   string
	Demandes   []DemandePersonne
}

type DemandePersonne struct {
	Demande     fs.Demande
	DemandeFile logic.PublicFile
	Uploaded    []logic.PublicFile
}

func (ct *Controller) LoadDocuments(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	out, err := ct.markAndloadDocuments(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// also update the [LastLoadDocuments] dossier field
func (ct *Controller) markAndloadDocuments(id ds.IdDossier) (Documents, error) {
	dossier, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return Documents{}, err
	}

	// start by updating last seen so that
	// the notification are updated properly
	dossier.Dossier.LastLoadDocuments = time.Now()
	_, err = dossier.Dossier.Update(ct.db)
	if err != nil {
		return Documents{}, utils.SQLError(err)
	}

	out, err := loadDocuments(ct.db, ct.key, dossier)
	if err != nil {
		return Documents{}, err
	}

	return out, nil
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
		// TODO: Only ask for documents is the camp is marked as Ready

		item := FilesCamp{idCamp: camp.Id, Camp: camp.Label()}
		// other files
		for _, link := range byCamp[camp.Id] {
			if (link.IsLettre && camp.DocumentsToShow.LettreDirecteur) ||
				!link.IsLettre {
				item.Files = append(item.Files, logic.NewPublicFile(key, campFiles[link.IdFile]))
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
	// always asks vaccins
	demandeVaccin, err := filesAPI.DemandeVaccin(db)
	if err != nil {
		return Documents{}, err
	}

	idsDemandes := append(links2.IdDemandes(), demandeVaccin.Id)
	personnesFiles, demandes, err := filesAPI.LoadFilesPersonnes(db, key, idsDemandes, dossier.Participants.IdPersonnes()...)
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
			// wait for the validation
			if personne.IsTemp || part.Statut != cps.Inscrit {
				continue
			}

			// add demandes from the camp
			for _, link := range demandesByCamp[part.IdCamp] {
				demande := demandes[link.IdDemande]
				dp := DemandePersonne{
					Demande:  demande,
					Uploaded: personnesFiles[personne.Id][demande.Id],
				}
				if fi := demande.IdFile; fi.Valid {
					dp.DemandeFile = logic.NewPublicFile(key, demandesFiles[fi.Id])
				}
				item.Demandes = append(item.Demandes, dp)
			}

			if asksFichesanitaire(dossier, personne) {
				item.Demandes = append(item.Demandes, DemandePersonne{
					Demande:  demandeVaccin,
					Uploaded: personnesFiles[personne.Id][demandeVaccin.Id],
				})
			}
		}
		// do not include empty lists
		if len(item.Demandes) != 0 {
			out.FilesToUpload = append(out.FilesToUpload, item)
		}
	}

	out.Fiches, err = loadFichesanitaires(db, dossier)
	if err != nil {
		return Documents{}, err
	}

	// chartes
	for _, personne := range dossier.Personnes()[1:] {
		// only ask for inscrits older than 12,
		// with camp asking for charte
		showCharte := false
		for _, camp := range dossier.CampsFor(personne.Id) {
			if camp.DocumentsReady && camp.DocumentsToShow.CharteParticipant && camp.AgeDebutCamp(personne.DateNaissance) >= 12 {
				showCharte = true
			}
		}
		if !showCharte {
			continue
		}

		out.Chartes = append(out.Chartes, Charte{
			Id:       personne.Id,
			Personne: personne.PrenomN(),
			Accepted: personne.CharteAccepted.After(dossier.Dossier.MomentInscription),
		})
	}

	out.setCounts(dossier.Dossier.LastLoadDocuments, dossier.Events)
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
) (logic.PublicFile, error) {
	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return logic.PublicFile{}, err
	}
	// basic security check
	if hasPersonne := slices.Contains(dossier.Participants.IdPersonnes(), idPersonne); !hasPersonne {
		return logic.PublicFile{}, errors.New("access forbidden")
	}
	file, err := filesAPI.SaveFileFor(ct.files, ct.db, idPersonne, idDemande, content, filename)
	if err != nil {
		return logic.PublicFile{}, err
	}
	return logic.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DeleteDocument(c echo.Context) error {
	key := c.QueryParam("key")
	err := filesAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) AccepteCharte(c echo.Context) error {
	token := c.QueryParam("token")
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	idPersonne, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	err = ct.accepteCharte(idDossier, idPersonne)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) accepteCharte(idDossier ds.IdDossier, idPersonne pr.IdPersonne) error {
	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	// check Id is valid
	if !slices.Contains(dossier.Participants.IdPersonnes(), idPersonne) {
		return errors.New("access forbidden")
	}
	personne, err := pr.SelectPersonne(ct.db, idPersonne)
	if err != nil {
		return utils.SQLError(err)
	}
	personne.CharteAccepted = time.Now()
	_, err = personne.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}
