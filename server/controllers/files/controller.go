package files

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/url"
	"path"
	"strconv"
	"time"

	"registro/assets"
	"registro/crypto"
	"registro/sql/files"
	fs "registro/sql/files"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// Controller exposes a global API for files,
// to read, update and delete files.
type Controller struct {
	db *sql.DB

	key   crypto.Encrypter
	files fs.FileSystem
}

func NewController(db *sql.DB, key crypto.Encrypter, files fs.FileSystem) *Controller {
	return &Controller{db, key, files}
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
	return sendBlob(c, content, file.NomClient)
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

func sendBlob(c echo.Context, content []byte, name string) error {
	mimeType := mime.TypeByExtension(path.Ext(name))
	u := url.URL{Path: name}
	name = u.String()
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+name)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(content)))
	return c.Blob(200, mimeType, content)
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

// PublicFile expose un accès protégé à un fichier,
// permettant téléchargement/suppression/modification.
type PublicFile struct {
	Id string // crypted

	// En bytes
	Taille    int
	NomClient string
	Uploaded  time.Time
}

func NewPublicFile(key crypto.Encrypter, file files.File) PublicFile {
	return PublicFile{
		Id:        crypto.EncryptID(key, file.Id),
		Taille:    file.Taille,
		NomClient: file.NomClient,
		Uploaded:  file.Uploaded,
	}
}
