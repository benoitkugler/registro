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

// StripeKey is a private key used to secure Stripe notifications
type StripeKey string

// NewStripe loads STRIPE_SECRET and STRIPE_WEBHOOK env. variables
func NewStripe() (string, StripeKey, error) {
	secret := os.Getenv("STRIPE_SECRET")
	if secret == "" {
		return "", "", errors.New("missing env STRIPE_SECRET")
	}

	webhook := os.Getenv("STRIPE_WEBHOOK")
	if webhook == "" {
		return "", "", errors.New("missing env STRIPE_WEBHOOK")
	}

	return secret, StripeKey(webhook), nil
}

type Helloasso struct {
	ID     string
	Secret string
}

// NewHelloasso uses env variables to build Helloasso credentials :
// HELLOASSO_ID, HELLOASSO_SECRET
func NewHelloasso() (out Helloasso, err error) {
	out.ID = os.Getenv("HELLOASSO_ID")
	if out.ID == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_ID")
	}

	out.Secret = os.Getenv("HELLOASSO_SECRET")
	if out.Secret == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_SECRET")
	}

	return out, nil
}
