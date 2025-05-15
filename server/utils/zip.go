package utils

import (
	"archive/zip"
	"fmt"
	"io"
)

// Zip is a builder for a .zip file
type Zip struct {
	z *zip.Writer
}

// NewZip builds a .zip archive into [dst]
func NewZip(dst io.Writer) *Zip {
	return &Zip{z: zip.NewWriter(dst)}
}

// AddFile ajoute un fichier.
func (a *Zip) AddFile(name string, content io.Reader) error {
	w, err := a.z.Create(name)
	if err != nil {
		return fmt.Errorf("compressing file: %s", err)
	}
	if _, err := io.Copy(w, content); err != nil {
		return fmt.Errorf("compressing file: %s", err)
	}
	if err := a.z.Flush(); err != nil {
		return fmt.Errorf("compressing file: %s", err)
	}
	return nil
}

// Close termine la création de l'archive.
func (a Zip) Close() error { return a.z.Close() }
