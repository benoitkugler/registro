package files

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"registro/utils"
)

// filepath returns the file location
func (f IdFile) filepath(dir string, forMiniature bool) string {
	id := fmt.Sprintf("file_%d", f)
	if forMiniature {
		id += "_min"
	}
	return filepath.Join(dir, id)
}

// FileSystem controle l'accès au contenu
// des fichiers (et leurs miniatures)
type FileSystem struct {
	root string
}

// [root] est le dossier dans lequel les fichiers sont stockés
func NewFileSystem(root string) FileSystem { return FileSystem{root: root} }

// Delete supprime le contenu et la miniature des fichiers donnés.
func (fs FileSystem) Delete(ids ...IdFile) error {
	for _, doc := range ids {
		filepath := doc.filepath(fs.root, false)
		err := os.Remove(filepath)
		if err != nil {
			return fmt.Errorf("failed to remove document (ID %d) : %s", doc, err)
		}

		filepath = doc.filepath(fs.root, true)
		err = os.Remove(filepath)
		if err != nil {
			return fmt.Errorf("failed to remove document miniature (ID %d) : %s", doc, err)
		}
	}
	return nil
}

func (fs FileSystem) Load(id IdFile, miniature bool) ([]byte, error) {
	filepath := id.filepath(fs.root, miniature)
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to load document (ID %d) : %s", id, err)
	}
	return content, nil
}

func (fs FileSystem) save(doc IdFile, fileContent []byte, miniature bool) error {
	filepath := doc.filepath(fs.root, miniature)
	err := os.WriteFile(filepath, fileContent, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save document (ID %d) : %s", doc, err)
	}
	return nil
}

// UploadFile computes the miniature, stores the content on the file system
// and updates the metadata.
// The file size is not checked.
func UploadFile(fs FileSystem, db DB, id IdFile, fileContent []byte, filename string) (File, error) {
	ext := filepath.Ext(filename)
	minContent, err := computeMiniature(ext, bytes.NewReader(fileContent))
	if err != nil {
		return File{}, err
	}
	err = fs.save(id, fileContent, false)
	if err != nil {
		return File{}, err
	}
	err = fs.save(id, minContent, true)
	if err != nil {
		return File{}, err
	}
	meta := File{Id: id, Taille: len(fileContent), NomClient: filename, Uploaded: time.Now().Truncate(time.Second)}
	meta, err = meta.Update(db)
	if err != nil {
		return File{}, utils.SQLError(err)
	}
	return meta, nil
}
