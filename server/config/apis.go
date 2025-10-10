package config

import (
	"errors"
	"os"
)

type Joomeo struct {
	Apikey, Login, Password string
	// the Joomeo top-level folder Label to use
	// to store sejours albums
	RootFolder string
}

// NewJoomeo uses env variables to build Joomeo credentials :
// JOOMEO_APIKEY, JOOMEO_LOGIN, JOOMEO_PASSWORD, JOOMEO_ROOT_FOLDER
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
	out.RootFolder = os.Getenv("JOOMEO_ROOT_FOLDER")
	if out.RootFolder == "" {
		return Joomeo{}, errors.New("missing env JOOMEO_ROOT_FOLDER")
	}

	return out, nil
}

// See https://dashboard.stripe.com/account/apikeys
type Stripe struct {
	Key string
	// Webhook is a private key used to secure Stripe notifications
	Webhook string
}

// NewStripe loads STRIPE_KEY and STRIPE_WEBHOOK env. variables
func NewStripe() (Stripe, error) {
	key := os.Getenv("STRIPE_KEY")
	if key == "" {
		return Stripe{}, errors.New("missing env STRIPE_KEY")
	}

	webhook := os.Getenv("STRIPE_WEBHOOK")
	if webhook == "" {
		return Stripe{}, errors.New("missing env STRIPE_WEBHOOK")
	}

	return Stripe{key, webhook}, nil
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
