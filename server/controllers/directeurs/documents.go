package directeurs

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"

	filesAPI "registro/controllers/files"
	fsAPI "registro/controllers/files"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

var errNoDir = errors.New("Aucun directeur n'est déclaré pour ce camp !")

type DocumentsOut struct {
	Ready  bool
	ToShow cps.DocumentsToShow
	// à télécharger (n'inclut pas la lettre)
	FilesToDownload []filesAPI.PublicFile
	CampDemandes    []DemandeDirecteur
	// indique si une lettre a été généré (au format PDF)
	HasLettre bool
	// public and private
	AvailableDemandes []DemandeDirecteur
}

type DemandeDirecteur struct {
	Demande fs.Demande
	File    filesAPI.PublicFile // valid if Demande.IdFile is non zero
}

func (ct *Controller) DocumentsGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getDocuments(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getDocuments(id cps.IdCamp) (DocumentsOut, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}
	directeur, _, err := ct.findDirecteur(camp.Id)
	if err != nil {
		return DocumentsOut{}, err
	}

	toDownload, err := fs.SelectFileCampsByIdCamps(ct.db, id)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}
	allDemandes, err := fs.SelectAllDemandes(ct.db)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}

	links, err := fs.SelectDemandeCampsByIdCamps(ct.db, id)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}
	appliedDemandes, err := fs.SelectDemandes(ct.db, links.IdDemandes()...)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}
	filesIDs := toDownload.IdFiles()
	for _, demande := range appliedDemandes {
		if file := demande.IdFile; file.Valid {
			filesIDs = append(filesIDs, file.Id)
		}
	}

	files, err := fs.SelectFiles(ct.db, filesIDs...)
	if err != nil {
		return DocumentsOut{}, utils.SQLError(err)
	}

	out := DocumentsOut{
		Ready:  camp.DocumentsReady,
		ToShow: camp.DocumentsToShow,
	}
	for _, link := range toDownload {
		if link.IsLettre {
			out.HasLettre = true
			// special case, not included in the list
			continue
		}
		out.FilesToDownload = append(out.FilesToDownload, filesAPI.NewPublicFile(ct.key, files[link.IdFile]))
	}
	for _, demande := range appliedDemandes {
		item := DemandeDirecteur{Demande: demande}
		if file := demande.IdFile; file.Valid {
			item.File = filesAPI.NewPublicFile(ct.key, files[file.Id])
		}
		out.CampDemandes = append(out.CampDemandes, item)
	}
	for _, demande := range allDemandes {
		if demande.Categorie != 0 { // only custom
			continue
		}
		if owner := demande.IdDirecteur; owner.Valid && owner.Id != directeur.Id {
			// private to someone else
			continue
		}
		item := DemandeDirecteur{Demande: demande}
		if file := demande.IdFile; file.Valid {
			item.File = filesAPI.NewPublicFile(ct.key, files[file.Id])
		}
		out.AvailableDemandes = append(out.AvailableDemandes, item)
	}

	slices.SortFunc(out.FilesToDownload, func(a, b filesAPI.PublicFile) int { return int(a.Id - b.Id) })
	slices.SortFunc(out.CampDemandes, func(a, b DemandeDirecteur) int { return int(a.Demande.Id - b.Demande.Id) })
	slices.SortFunc(out.AvailableDemandes, func(a, b DemandeDirecteur) int { return int(a.Demande.Id - b.Demande.Id) })

	return out, nil
}

func (ct *Controller) DocumentsUpdateToShow(c echo.Context) error {
	user := JWTUser(c)

	var args cps.DocumentsToShow
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateToShow(user, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateToShow(id cps.IdCamp, args cps.DocumentsToShow) error {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	camp.DocumentsToShow = args
	_, err = camp.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) DocumentsUploadToDownload(c echo.Context) error {
	user := JWTUser(c)

	header, err := c.FormFile("document")
	if err != nil {
		return err
	}
	content, filename, err := filesAPI.ReadUpload(header)
	if err != nil {
		return err
	}
	out, err := ct.uploadToDownload(user, content, filename)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) uploadToDownload(idCamp cps.IdCamp, content []byte, filename string) (filesAPI.PublicFile, error) {
	var (
		file fs.File
		err  error
	)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// create a new file, and the associated metadata
		file, err = fs.File{}.Insert(tx)
		if err != nil {
			return err
		}
		err = fs.FileCamp{IdFile: file.Id, IdCamp: idCamp, IsLettre: false}.Insert(tx)
		if err != nil {
			return err
		}
		file, err = fs.UploadFile(ct.files, tx, file.Id, content, filename)
		if err != nil {
			return err
		}
		return nil
	})

	return filesAPI.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DocumentsDeleteToDownload(c echo.Context) error {
	key := c.QueryParam("key")
	err := filesAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// DocumentsApplyDemande applique une demande déjà existante au séjour
func (ct *Controller) DocumentsApplyDemande(c echo.Context) error {
	user := JWTUser(c)
	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	out, err := ct.applyDemande(user, idDemande)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) applyDemande(idCamp cps.IdCamp, idDemande fs.IdDemande) (DemandeDirecteur, error) {
	demande, err := fs.SelectDemande(ct.db, idDemande)
	if err != nil {
		return DemandeDirecteur{}, utils.SQLError(err)
	}
	// TODO: we should check if the directeur is allowed to use
	// this demande
	out := DemandeDirecteur{Demande: demande}
	if fi := demande.IdFile; fi.Valid {
		file, err := fs.SelectFile(ct.db, fi.Id)
		if err != nil {
			return DemandeDirecteur{}, utils.SQLError(err)
		}
		out.File = filesAPI.NewPublicFile(ct.key, file)
	}
	err = fs.DemandeCamp{IdCamp: idCamp, IdDemande: idDemande}.Insert(ct.db)
	if err != nil {
		return DemandeDirecteur{}, utils.SQLError(err)
	}
	return out, nil
}

// DocumentsUnapplyDemande remove the link object, not the demande it self
func (ct *Controller) DocumentsUnapplyDemande(c echo.Context) error {
	user := JWTUser(c)
	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	err = ct.unapplyDemande(user, idDemande)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) unapplyDemande(id cps.IdCamp, idDemande fs.IdDemande) error {
	err := fs.DemandeCamp{IdCamp: id, IdDemande: idDemande}.Delete(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

// DocumentsCreateDemande creates a new [Demande], linked to
// the directeur.
// It returns an error if there is no directeur.
func (ct *Controller) DocumentsCreateDemande(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.createDemande(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createDemande(id cps.IdCamp) (DemandeDirecteur, error) {
	directeur, hasDirecteur, err := ct.findDirecteur(id)
	if err != nil {
		return DemandeDirecteur{}, err
	}
	if !hasDirecteur {
		return DemandeDirecteur{}, errNoDir
	}
	demande, err := fs.Demande{
		IdDirecteur: directeur.Id.Opt(),
		Categorie:   fs.NoBuiltin,
		MaxDocs:     1,
		JoursValide: 0,
	}.Insert(ct.db)
	if err != nil {
		return DemandeDirecteur{}, utils.SQLError(err)
	}
	return DemandeDirecteur{Demande: demande}, nil
}

func (ct *Controller) DocumentsUpdateDemande(c echo.Context) error {
	user := JWTUser(c)
	var args fs.Demande
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateDemande(user, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) checkDemandeOwner(idCamp cps.IdCamp, idDemande fs.IdDemande) (fs.Demande, error) {
	dir, hasDirecteur, err := ct.findDirecteur(idCamp)
	if err != nil {
		return fs.Demande{}, err
	}
	if !hasDirecteur {
		return fs.Demande{}, errNoDir
	}
	demande, err := fs.SelectDemande(ct.db, idDemande)
	if err != nil {
		return fs.Demande{}, utils.SQLError(err)
	}
	// only allow updates on private, owned [Demande]s
	if !demande.IdDirecteur.Is(dir.Id) {
		return fs.Demande{}, errors.New("access forbidden")
	}
	return demande, nil
}

func (ct *Controller) updateDemande(id cps.IdCamp, args fs.Demande) error {
	demande, err := ct.checkDemandeOwner(id, args.Id)
	if err != nil {
		return err
	}
	demande.Description = args.Description
	demande.MaxDocs = args.MaxDocs
	demande.JoursValide = args.JoursValide
	_, err = demande.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

// DocumentsDeleteDemande deletes the Demande and all associated files,
// if any.
func (ct *Controller) DocumentsDeleteDemande(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	err = ct.deleteDemande(user, id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteDemande(user cps.IdCamp, id fs.IdDemande) error {
	demande, err := ct.checkDemandeOwner(user, id)
	if err != nil {
		return err
	}
	links, err := fs.SelectFilePersonnesByIdDemandes(ct.db, demande.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	filesToDelete := links.IdFiles()
	if file := demande.IdFile; file.Valid {
		filesToDelete = append(filesToDelete, file.Id)
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = fs.DeleteFilesByIDs(tx, filesToDelete...)
		if err != nil {
			return err
		}
		_, err = fs.DeleteDemandeById(tx, id)
		if err != nil {
			return err
		}
		err = ct.files.Delete(filesToDelete...) // contenu
		if err != nil {
			return err
		}
		return nil
	})
}

func (ct *Controller) DocumentsUploadDemandeFile(c echo.Context) error {
	user := JWTUser(c)

	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	header, err := c.FormFile("document")
	if err != nil {
		return err
	}
	content, filename, err := filesAPI.ReadUpload(header)
	if err != nil {
		return err
	}
	out, err := ct.uploadDemandeFile(user, idDemande, content, filename)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) uploadDemandeFile(user cps.IdCamp, idDemande fs.IdDemande, content []byte, filename string) (filesAPI.PublicFile, error) {
	demande, err := ct.checkDemandeOwner(user, idDemande)
	if err != nil {
		return filesAPI.PublicFile{}, err
	}
	var file fs.File
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// create a new file, and the associated metadata
		file, err = fs.File{}.Insert(tx)
		if err != nil {
			return err
		}
		demande.IdFile = file.Id.Opt()
		_, err = demande.Update(tx)
		if err != nil {
			return err
		}
		file, err = fs.UploadFile(ct.files, tx, file.Id, content, filename)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return filesAPI.PublicFile{}, err
	}

	return filesAPI.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DocumentsDeleteDemandeFile(c echo.Context) error {
	key := c.QueryParam("key")
	err := filesAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// download API

type DocumentsUploadedOut struct {
	Personnes         pr.Personnes
	DemandesDocuments []DemandeDocuments
}

type DemandeDocuments struct {
	Demande    fs.Demande
	UploadedBy []pr.IdPersonne // the ones with a file
}

// DocumentsGetUploaded renvoie les demandes et fichiers ajoutés par
// les participants.
func (ct *Controller) DocumentsGetUploaded(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getUploaded(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getUploaded(id cps.IdCamp) (DocumentsUploadedOut, error) {
	// demandes
	links1, err := fs.SelectDemandeCampsByIdCamps(ct.db, id)
	if err != nil {
		return DocumentsUploadedOut{}, utils.SQLError(err)
	}
	demandes, err := fs.SelectDemandes(ct.db, links1.IdDemandes()...)
	if err != nil {
		return DocumentsUploadedOut{}, utils.SQLError(err)
	}
	// personnes et fichiers
	camp, err := cps.LoadCampPersonnes(ct.db, id)
	if err != nil {
		return DocumentsUploadedOut{}, err
	}
	personnes := camp.Personnes(true)
	links2, err := fs.SelectFilePersonnesByIdPersonnes(ct.db, personnes.IDs()...)
	if err != nil {
		return DocumentsUploadedOut{}, utils.SQLError(err)
	}
	filesByDemande := links2.ByIdDemande()

	out := DocumentsUploadedOut{Personnes: personnes}
	for _, demande := range demandes {
		out.DemandesDocuments = append(out.DemandesDocuments, DemandeDocuments{
			Demande:    demande,
			UploadedBy: filesByDemande[demande.Id].IdPersonnes(),
		})
	}

	slices.SortFunc(out.DemandesDocuments, func(a, b DemandeDocuments) int { return int(a.Demande.Id - b.Demande.Id) })

	return out, nil
}

// DocumentsStreamUploaded télécharge tous les fichiers pour une [Demande],
// dans une archive .ZIP
func (ct *Controller) DocumentsStreamUploaded(c echo.Context) error {
	user := JWTUser(c)
	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	files, archiveName, err := ct.selectDocumentsForDemande(user, idDemande)
	if err != nil {
		return err
	}
	return fsAPI.StreamZip(c.Response(), archiveName, func(yield func(fsAPI.ZipItem, error) bool) {
		for _, file := range files {
			content, err := ct.files.Load(file.Id, false)
			if err != nil {
				yield(fsAPI.ZipItem{}, err)
				return
			}
			if !yield(fsAPI.ZipItem{Name: file.NomClient, Content: content}, nil) {
				return
			}
		}
	})
}

func (ct *Controller) selectDocumentsForDemande(idCamp cps.IdCamp, idDemande fs.IdDemande) (fs.Files, string, error) {
	// personnes et fichiers
	camp, err := cps.LoadCampPersonnes(ct.db, idCamp)
	if err != nil {
		return nil, "", err
	}
	personnes := camp.Personnes(true)
	links, err := fs.SelectFilePersonnesByIdPersonnes(ct.db, personnes.IDs()...)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	files, err := fs.SelectFiles(ct.db, links.ByIdDemande()[idDemande].IdFiles()...)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	archiveName := fmt.Sprintf("Fichiers %s.zip", camp.Camp.Label())
	return files, archiveName, nil
}
