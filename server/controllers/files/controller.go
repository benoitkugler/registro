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
	"strconv"

	"registro/assets"
	"registro/config"
	"registro/crypto"
	"registro/logic"
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

func (ct *Controller) Get(c echo.Context) error {
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

// GetMiniature returns a placeholder image on error
func (ct *Controller) GetMiniature(c echo.Context) error {
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
	resp.Header().Set(echo.HeaderContentType, "application/x-zip")
	resp.Header().Set(echo.HeaderContentDisposition, utils.AttachementHeader(archiveName))
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
func Delete(db *sql.DB, enc crypto.Encrypter, files fs.FileSystem, key string) error {
	id, err := crypto.DecryptID[fs.IdFile](enc, key)
	if err != nil {
		return err
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
		return err
	}
	return nil
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

func LoadVaccins(db fs.DB, key crypto.Encrypter, personnes []pr.IdPersonne) (map[pr.IdPersonne][]logic.PublicFile, fs.Demande, error) {
	vaccinDemande, err := DemandeVaccin(db)
	if err != nil {
		return nil, fs.Demande{}, err
	}

	files, _, err := LoadFilesPersonnes(db, key, []fs.IdDemande{vaccinDemande.Id}, personnes...)
	if err != nil {
		return nil, fs.Demande{}, err
	}
	return files[vaccinDemande.Id], vaccinDemande, nil
}

func LoadFilesPersonnes(db fs.DB, key crypto.Encrypter, demandes []fs.IdDemande, personnes ...pr.IdPersonne) (map[fs.IdDemande]map[pr.IdPersonne][]logic.PublicFile, fs.Demandes,
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
	byDemande := tmp.ByIdDemande()

	out := make(map[fs.IdDemande]map[pr.IdPersonne][]logic.PublicFile)
	for _, idDemande := range demandes {
		links := byDemande[idDemande].ByIdPersonne()
		demandes := make(map[pr.IdPersonne][]logic.PublicFile, len(links))
		for idPersonne, innerLinks := range links {
			files := make([]logic.PublicFile, len(innerLinks))
			for i, file := range innerLinks {
				files[i] = logic.NewPublicFile(key, allFiles[file.IdFile])
			}
			demandes[idPersonne] = files
		}
		out[idDemande] = demandes
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
