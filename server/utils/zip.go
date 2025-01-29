package utils

import (
	"archive/zip"
	"bytes"
	"io"
)

// Zip is a builder for a .zip file
type Zip struct {
	b *bytes.Buffer
	z *zip.Writer

	err error // erreur courante
}

func NewZip() *Zip {
	buf := new(bytes.Buffer)
	return &Zip{b: buf, z: zip.NewWriter(buf)}
}

// AddFile ajoute un fichier. La gestion de l'erreur est reporté à
// la méthode `Close`.
func (a *Zip) AddFile(name string, content io.Reader) {
	if a.err != nil {
		return
	}
	w, err := a.z.Create(name)
	if err != nil {
		a.err = err
		return
	}
	if _, err := io.Copy(w, content); err != nil {
		a.err = err
	}
}

// Close termine la création de l'archive et renvoie son contenu.
func (a Zip) Close() (*bytes.Buffer, error) {
	if a.err != nil {
		return nil, a.err
	}
	if err := a.z.Close(); err != nil {
		return nil, err
	}
	return a.b, nil
}
