package config

import (
	"errors"
	"os"
)

// Keys expose the passwords/keys used to
// authenticate/crypt
type Keys struct {
	EncryptKey  string // used for encryption key
	Backoffice  string // password
	FondSoutien string // password
	Directeurs  string // global password
}

// NewKeys uses env. variables to load the credentials :
// SERVER_KEY, BACKOFFICE_PASSWORD, DIRECTEURS_PASSWORD
func NewKeys() (keys Keys, _ error) {
	keys.EncryptKey = os.Getenv("SERVER_KEY")
	if keys.EncryptKey == "" {
		return keys, errors.New("missing env. SERVER_KEY (encryption key)")
	}
	keys.Backoffice = os.Getenv("BACKOFFICE_PASSWORD")
	if keys.Backoffice == "" {
		return keys, errors.New("missing env. BACKOFFICE_PASSWORD")
	}
	keys.FondSoutien = os.Getenv("FOND_SOUTIEN_PASSWORD")
	if keys.FondSoutien == "" {
		return keys, errors.New("missing env. FOND_SOUTIEN_PASSWORD")
	}
	keys.Directeurs = os.Getenv("DIRECTEURS_PASSWORD")
	if keys.Directeurs == "" {
		return keys, errors.New("missing env. DIRECTEURS_PASSWORD")
	}
	return keys, nil
}
