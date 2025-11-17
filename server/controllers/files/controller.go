package files

import (
	"bytes"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"iter"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	"registro/assets"
	"registro/config"
	"registro/crypto"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// Controller exposes a global API for files,
// to read, update and delete files.
type Controller struct {
	db *sql.DB

	key   crypto.Encrypter
	files fs.FileSystem
	asso  config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, files fs.FileSystem, asso config.Asso) *Controller {
	return &Controller{db, key, files, asso}
}

func (ct *Controller) LoadDocument(c echo.Context) error {
	key := c.QueryParam("key")
	id, err := crypto.DecryptID[fs.IdFile](ct.key, key)
	if err != nil {
		return err
	}
	file, err := fs.SelectFile(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	content, err := ct.files.Load(id, false)
	if err != nil {
		return err
	}
	mimeType := SetBlobHeader(c, content, file.NomClient)
	return c.Blob(200, mimeType, content)
}

// LoadMiniature returns a placeholder image on error
func (ct *Controller) LoadMiniature(c echo.Context) error {
	key := c.QueryParam("key")
	id, err := crypto.DecryptID[fs.IdFile](ct.key, key)
	if err != nil {
		log.Println(err)
		return c.Blob(200, mime.TypeByExtension(".png"), assets.DefaultMiniaturePNG)
	}
	content, err := ct.files.Load(id, true)
	if err != nil {
		log.Println(err)
		return c.Blob(200, mime.TypeByExtension(".png"), assets.DefaultMiniaturePNG)
	}
	return c.Blob(200, mime.TypeByExtension(".png"), content)
}

// SetBlobHeader sets Content-Disposition and Content-Length headers
// and returns the mime type
func SetBlobHeader(c echo.Context, content []byte, name string) string {
	u := url.URL{Path: name}
	name = u.String()
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+name)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(content)))
	return mime.TypeByExtension(path.Ext(name))
}

type ZipItem struct {
	Name    string
	Content []byte
}

// StreamZip writes [files] into a .ZIP response,
// properly escaping [archiveName].
func StreamZip(resp http.ResponseWriter, archiveName string, files iter.Seq2[ZipItem, error]) error {
	resp.Header().Set("Content-Type", "application/x-zip")
	resp.Header().Set("Content-Disposition", utils.AttachementHeader(archiveName))
	if flusher, ok := resp.(http.Flusher); ok {
		// this is needed so that browsers display a progress bar
		flusher.Flush()
	}

	archive := utils.NewZip(resp)

	for file, err := range files {
		if err != nil {
			return err
		}
		err := archive.AddFile(file.Name, bytes.NewReader(file.Content))
		if err != nil {
			return err
		}
		if flusher, ok := resp.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	err := archive.Close()
	if err != nil {
		return err
	}

	return nil
}

// Delete removes the file identified by the crypted ID [key]
func Delete(db *sql.DB, enc crypto.Encrypter, files fs.FileSystem, key string) (fs.IdFile, error) {
	id, err := crypto.DecryptID[fs.IdFile](enc, key)
	if err != nil {
		return 0, err
	}
	err = utils.InTx(db, func(tx *sql.Tx) error {
		_, err = fs.DeleteFileById(tx, id)
		if err != nil {
			return err
		}
		err = files.Delete(id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return id, err
	}
	return id, nil
}

// ReadUpload checks the file size and reads its content.
// The size of the file is checked against the max 5MB
func ReadUpload(fileHeader *multipart.FileHeader) (content []byte, filename string, err error) {
	const MB = 1000000
	const maxSize = 5 * MB
	if fileHeader.Size > maxSize {
		return nil, "", fmt.Errorf("file too large (%d MB)", fileHeader.Size/MB)
	}

	f, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	content, err = io.ReadAll(f)
	if err != nil {
		return nil, "", err
	}
	if int64(len(content)) != fileHeader.Size {
		return nil, "", errors.New("invalid file size")
	}
	return content, fileHeader.Filename, nil
}

func DemandeVaccin(db fs.DB) (fs.Demande, error) {
	demandes, err := fs.SelectAllDemandes(db)
	if err != nil {
		return fs.Demande{}, utils.SQLError(err)
	}
	for _, demande := range demandes {
		if demande.Categorie == fs.Vaccins {
			return demande, nil
		}
	}
	return fs.Demande{}, errors.New("missing Demande for categorie <Vaccins>")
}

func LoadFilesPersonnes(db fs.DB, key crypto.Encrypter, demandes []fs.IdDemande, personnes ...pr.IdPersonne) (map[pr.IdPersonne]map[fs.IdDemande][]logic.PublicFile, fs.Demandes,
	error,
) {
	demandesM, err := fs.SelectDemandes(db, demandes...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	tmp, err := fs.SelectFilePersonnesByIdPersonnes(db, personnes...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	allFiles, err := fs.SelectFiles(db, tmp.IdFiles()...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}
	byPersonne := tmp.ByIdPersonne()

	out := make(map[pr.IdPersonne]map[fs.IdDemande][]logic.PublicFile)
	for idPersonne, links := range byPersonne {
		byDemande := links.ByIdDemande()
		demandesM := make(map[fs.IdDemande][]logic.PublicFile, len(links))
		for _, idDemande := range demandes {
			filesLink := byDemande[idDemande]
			files := make([]logic.PublicFile, len(filesLink))
			for i, file := range filesLink {
				files[i] = logic.NewPublicFile(key, allFiles[file.IdFile])
			}
			demandesM[idDemande] = files
		}
		out[idPersonne] = demandesM
	}
	return out, demandesM, nil
}

func SaveFileFor(files fs.FileSystem, db *sql.DB, idPersonne pr.IdPersonne, idDemande fs.IdDemande, content []byte, filename string) (file fs.File, err error) {
	err = utils.InTx(db, func(tx *sql.Tx) error {
		// create a new file, and the associated metadata
		file, err = fs.File{}.Insert(tx)
		if err != nil {
			return err
		}
		err = fs.FilePersonne{IdFile: file.Id, IdPersonne: idPersonne, IdDemande: idDemande}.Insert(tx)
		if err != nil {
			return err
		}
		file, err = fs.UploadFile(files, tx, file.Id, content, filename)
		if err != nil {
			return err
		}
		return nil
	})
	return file, err
}

// documents

type ParticipantsFilesLoader struct {
	camps    cps.CampsData
	dossiers ds.Dossiers

	demandes       fs.Demandes
	idVaccin       fs.IdDemande
	demandesByCamp map[cps.IdCamp]fs.DemandeCamps

	fiches map[pr.IdPersonne]pr.Fichesanitaire
	files  map[pr.IdPersonne]map[fs.IdDemande][]logic.PublicFile
}

func LoadParticipantsFiles(db fs.DB, key crypto.Encrypter, ids []cps.IdCamp) (ParticipantsFilesLoader, error) {
	camps, err := cps.LoadCamps(db, ids)
	if err != nil {
		return ParticipantsFilesLoader{}, err
	}

	dossiers, err := ds.SelectDossiers(db, camps.IdDossiers()...)
	if err != nil {
		return ParticipantsFilesLoader{}, utils.SQLError(err)
	}

	// demandes
	// always include vaccin
	vaccinDemande, err := DemandeVaccin(db)
	if err != nil {
		return ParticipantsFilesLoader{}, err
	}
	tmp, err := fs.SelectDemandeCampsByIdCamps(db, ids...)
	if err != nil {
		return ParticipantsFilesLoader{}, utils.SQLError(err)
	}
	idDemandes := append(tmp.IdDemandes(), vaccinDemande.Id)

	// personnes et fichiers
	personnes := camps.Personnes(true)
	tmp2, err := pr.SelectFichesanitairesByIdPersonnes(db, personnes.IDs()...)
	if err != nil {
		return ParticipantsFilesLoader{}, utils.SQLError(err)
	}
	fiches := tmp2.ByIdPersonne()

	files, demandes, err := LoadFilesPersonnes(db, key, idDemandes, personnes.IDs()...)
	if err != nil {
		return ParticipantsFilesLoader{}, err
	}

	return ParticipantsFilesLoader{camps, dossiers, demandes, vaccinDemande.Id, tmp.ByIdCamp(), fiches, files}, nil
}

type ParticipantFiles struct {
	Id             cps.IdParticipant
	Personne       string
	Fichesanitaire pr.FichesanitaireState
	Files          map[fs.IdDemande][]logic.PublicFile
}

// ParticipantsFiles is a 2D array participants as rows
// and [Demande]s as columns
type ParticipantsFiles struct {
	DocumentsReady bool // copied from Camp

	// contains Vaccins, and a column for Fiche sanitaire should be added
	Demandes     fs.Demandes
	Participants []ParticipantFiles
}

func (ld ParticipantsFilesLoader) For(id cps.IdCamp) ParticipantsFiles {
	idDemandes := append(ld.demandesByCamp[id].IdDemandes(), ld.idVaccin)
	demandes := make(fs.Demandes)
	for _, id := range idDemandes {
		demandes[id] = ld.demandes[id]
	}
	camp := ld.camps.For(id)

	out := ParticipantsFiles{DocumentsReady: camp.Camp.DocumentsReady, Demandes: demandes}
	for _, participant := range camp.Participants(true) {
		personne, dossier := participant.Personne, ld.dossiers[participant.Participant.IdDossier]
		filesM := make(map[fs.IdDemande][]logic.PublicFile)
		for _, demande := range idDemandes {
			filesM[demande] = ld.files[personne.Id][demande]
		}
		out.Participants = append(out.Participants, ParticipantFiles{
			Id:             participant.Participant.Id,
			Personne:       personne.NOMPrenom(),
			Fichesanitaire: ld.fiches[personne.Id].State(dossier.MomentInscription),
			Files:          filesM,
		})
	}

	slices.SortFunc(out.Participants, func(a, b ParticipantFiles) int { return strings.Compare(a.Personne, b.Personne) })

	return out
}

type DemandeStat struct {
	Title         string
	UploadedCount int
	InscritsCount int
}

// Stats returns a summary of the files uploaded
// for each [Demande] and Fiche Sanitaire
func (files ParticipantsFiles) Stats() (out []DemandeStat) {
	inscritsCount := len(files.Participants)

	// fiche sanitaire
	uploadedCount := 0
	for _, participant := range files.Participants {
		if hasFiche := participant.Fichesanitaire == pr.UpToDate; hasFiche {
			uploadedCount += 1
		}
	}
	out = append(out, DemandeStat{Title: "Fiches sanitaires", UploadedCount: uploadedCount, InscritsCount: inscritsCount})

	for _, demande := range files.Demandes {
		uploadedCount := 0
		for _, participant := range files.Participants {
			if hasFile := len(participant.Files[demande.Id]) > 0; hasFile {
				uploadedCount += 1
			}
		}
		out = append(out, DemandeStat{Title: demande.Title(), UploadedCount: uploadedCount, InscritsCount: inscritsCount})
	}

	return out
}
