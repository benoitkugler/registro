package logic

import (
	"time"

	"registro/crypto"
	"registro/sql/files"
)

// FilePublic expose un accès protégé à un fichier,
// permettant téléchargement/suppression/modification.
type FilePublic struct {
	Id string // crypted

	// En bytes
	Taille    int
	NomClient string
	Uploaded  time.Time
}

func PublishFile(key crypto.Encrypter, file files.File) FilePublic {
	return FilePublic{
		Id:        crypto.EncryptID(key, file.Id),
		Taille:    file.Taille,
		NomClient: file.NomClient,
		Uploaded:  file.Uploaded,
	}
}
