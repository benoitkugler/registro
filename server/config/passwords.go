package config

import (
	"errors"
	"os"
)

// Keys expose the passwords/keys used to
// authenticate/crypt
type Keys struct {
	ServerEnc  string // used for encryption key
	Backoffice string // password
	Directeurs string // global password
}

// NewKeys uses env. variables to load the credentials :
// SERVER_KEY, BACKOFFICE_PASSWORD, DIRECTEURS_PASSWORD
func NewKeys() (keys Keys, _ error) {
	keys.ServerEnc = os.Getenv("SERVER_KEY")
	if keys.ServerEnc == "" {
		return keys, errors.New("missing env. SERVER_KEY (encryption key)")
	}
	keys.Backoffice = os.Getenv("BACKOFFICE_PASSWORD")
	if keys.Backoffice == "" {
		return keys, errors.New("missing env. BACKOFFICE_PASSWORD")
	}
	keys.Directeurs = os.Getenv("DIRECTEURS_PASSWORD")
	if keys.Directeurs == "" {
		return keys, errors.New("missing env. DIRECTEURS_PASSWORD")
	}
	return keys, nil
}
