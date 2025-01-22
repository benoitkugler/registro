package config

import (
	"errors"
	"os"
)

type Joomeo struct {
	Apikey, Login, Password string
}

// NewJoomeo uses env variables to build Joomeo credentials :
// JOOMEO_APIKEY, JOOMEO_LOGIN, JOOMEO_PASSWORD
func NewJoomeo() (out Joomeo, err error) {
	out.Apikey = os.Getenv("JOOMEO_APIKEY")
	if out.Apikey == "" {
		return Joomeo{}, errors.New("missing env JOOMEO_APIKEY")
	}

	out.Login = os.Getenv("JOOMEO_LOGIN")
	if out.Login == "" {
		return Joomeo{}, errors.New("missing env JOOMEO_LOGIN")
	}

	out.Password = os.Getenv("JOOMEO_PASSWORD")
	if out.Password == "" {
		return Joomeo{}, errors.New("missing env JOOMEO_PASSWORD")
	}

	return out, nil
}
