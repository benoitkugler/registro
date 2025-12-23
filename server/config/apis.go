package config

import (
	"errors"
	"os"
)

// See https://api.immich.app
type Immich struct {
	// URL containing scheme and host, without path
	BaseURL string
	ApiKey  string
}

// NewImmich uses env variables to build Immich credentials :
// IMMICH_BASE_URL, IMMICH_API_KEY
func NewImmich() (out Immich, err error) {
	out.BaseURL = os.Getenv("IMMICH_BASE_URL")
	if out.BaseURL == "" {
		return Immich{}, errors.New("missing env IMMICH_BASE_URL")
	}

	out.ApiKey = os.Getenv("IMMICH_API_KEY")
	if out.ApiKey == "" {
		return Immich{}, errors.New("missing env IMMICH_API_KEY")
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
	Slug    string // asso slug
	Sandbox bool
	ID      string
	Secret  string
}

// NewHelloasso uses env variables to build Helloasso credentials :
// HELLOASSO_SLUG, HELLOASSO_SANDBOX, HELLOASSO_ID, HELLOASSO_SECRET
func NewHelloasso() (out Helloasso, err error) {
	out.Slug = os.Getenv("HELLOASSO_SLUG")
	if out.Slug == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_SLUG")
	}

	out.ID = os.Getenv("HELLOASSO_ID")
	if out.ID == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_ID")
	}

	out.Secret = os.Getenv("HELLOASSO_SECRET")
	if out.Secret == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_SECRET")
	}

	sandbox := os.Getenv("HELLOASSO_SANDBOX")
	if sandbox == "" {
		return Helloasso{}, errors.New("missing env HELLOASSO_SANDBOX")
	}
	out.Sandbox = sandbox == "TRUE"

	return out, nil
}
