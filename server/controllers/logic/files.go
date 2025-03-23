package logic

import (
	"time"

	"registro/crypto"
	"registro/sql/files"
)

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
