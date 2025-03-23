package files

import (
	"database/sql"
	"mime"
	"net/url"
	"path"
	"strconv"

	"registro/crypto"
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

func (ct *Controller) GetMiniature(c echo.Context) error {
	key := c.QueryParam("key")
	id, err := crypto.DecryptID[fs.IdFile](ct.key, key)
	if err != nil {
		return err
	}
	content, err := ct.files.Load(id, true)
	if err != nil {
		return err
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
