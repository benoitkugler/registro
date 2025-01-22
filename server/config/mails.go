package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// SMTP stores email credentials
type SMTP struct {
	Host, Password string

	// Should be a valid adress mail
	User string
	Port int

	// If false, uses dev adress to avoid
	// sending mail to "real" people
	Prod bool
}

// NewSMTP uses env variables to build SMTP credentials :
// SMTP_HOST, SMTP_USER, SMTP_PASSWORD, SMTP_PORT
func NewSMTP(prod bool) (out SMTP, err error) {
	out.Prod = prod

	out.Host = os.Getenv("SMTP_HOST")
	if out.Host == "" {
		return SMTP{}, errors.New("missing env SMTP_HOST")
	}

	out.User = os.Getenv("SMTP_USER")
	if out.User == "" {
		return SMTP{}, errors.New("missing env SMTP_USER")
	}

	out.Password = os.Getenv("SMTP_PASSWORD")
	if out.Password == "" {
		return SMTP{}, errors.New("missing env SMTP_PASSWORD")
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		return SMTP{}, errors.New("missing env SMTP_PORT")
	}
	out.Port, err = strconv.Atoi(port)
	if err != nil {
		return SMTP{}, fmt.Errorf("invalid SMTP_PORT: %s", err)
	}

	return out, nil
}
