package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// See https://api.immich.app
type Immich struct {
	// URL containing scheme and host, without path
	BaseURL url.URL
	ApiKey  string
}

// NewImmich uses env variables to build Immich credentials :
// IMMICH_BASE_URL, IMMICH_API_KEY
func NewImmich() (out Immich, err error) {
	baseURL := os.Getenv("IMMICH_BASE_URL")
	if baseURL == "" {
		return Immich{}, errors.New("missing env IMMICH_BASE_URL")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return out, fmt.Errorf("invalid env IMMICH_BASE_URL: %s", baseURL)
	}
	out.BaseURL = *u

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

// ANCV holds the credentials used to call the ANCV Chèque-Vacances Connect
// API. The association is connected directly, without an intermediary, so the
// merchant (PTL) sealing key is used.
//
// Kit d'intégration ANCV v1.08, §5.3.2.2 (p. 52) : « Si le paramètre
// Merchant → serviceProviderId est renseigné [...] la clé de signature à
// utiliser est celle de l'Intermédiaire de Service (IS). Dans le cas contraire,
// la clé de signature à utiliser est celle du commerçant (PTL). »
// §4.2 (p. 13) describes the two base URLs selected by [ANCV.Prod].
type ANCV struct {
	ShopID      int64  // merchant.shopId
	SealKey     string // merchant sealing key (PTL)
	SealVersion string // versionKey in the ANCV-Security header
	Prod        bool   // true => production, false => recette (safe zero value)
}

// NewANCV uses env variables to build ANCV credentials :
// ANCV_SHOP_ID, ANCV_SEAL_KEY, ANCV_SEAL_VERSION, ANCV_PROD (optional).
//
// Kit d'intégration ANCV v1.08, §4.3.1 (p. 13) : « L'en-tête doit
// contenir à minima les entrées suivantes : Content-Type : application/json,
// Ancv-Security : {algo.version.seal} ».
func NewANCV() (out ANCV, err error) {
	shopID := os.Getenv("ANCV_SHOP_ID")
	if shopID == "" {
		return ANCV{}, errors.New("missing env ANCV_SHOP_ID")
	}
	out.ShopID, err = strconv.ParseInt(shopID, 10, 64)
	if err != nil {
		return ANCV{}, fmt.Errorf("invalid env ANCV_SHOP_ID: %s", shopID)
	}

	out.SealKey = os.Getenv("ANCV_SEAL_KEY")
	if out.SealKey == "" {
		return ANCV{}, errors.New("missing env ANCV_SEAL_KEY")
	}

	out.SealVersion = os.Getenv("ANCV_SEAL_VERSION")
	if out.SealVersion == "" {
		return ANCV{}, errors.New("missing env ANCV_SEAL_VERSION")
	}

	// Prod is opt-in: an unset or unexpected value selects the recette API.
	out.Prod = os.Getenv("ANCV_PROD") == "TRUE"

	return out, nil
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

type Horizon struct {
	HorizonKey string // authentify incomming requests
	RelaisKey  string // authentify responses
}

// NewHorizon uses env variables to build Horizon credentials :
// HORIZON_HORIZON_KEY, HORIZON_RELAS_KEY
func NewHorizon() (out Horizon, err error) {
	out.HorizonKey = os.Getenv("HELLOASSO_HORIZON_KEY")
	if out.HorizonKey == "" {
		return Horizon{}, errors.New("missing env HELLOASSO_HORIZON_KEY")
	}

	out.RelaisKey = os.Getenv("HELLOASSO_RELAIS_KEY")
	if out.RelaisKey == "" {
		return Horizon{}, errors.New("missing env HELLOASSO_RELAIS_KEY")
	}

	return out, nil
}

type Payrexx struct {
	Instance string
	ApiKey   string
}

// NewPayrexx loads PAYREXX_INSTANCE and PAYREXX_API_KEY env. variables
func NewPayrexx() (Payrexx, error) {
	instance := os.Getenv("PAYREXX_INSTANCE")
	if instance == "" {
		return Payrexx{}, errors.New("missing env PAYREXX_INSTANCE")
	}

	apiKey := os.Getenv("PAYREXX_API_KEY")
	if apiKey == "" {
		return Payrexx{}, errors.New("missing env PAYREXX_API_KEY")
	}

	return Payrexx{instance, apiKey}, nil
}

type Compta struct {
	// Service account json file
	GoogleDriveKeyFile string
	// This folder must be shared with the service account
	// defined in [ComptaStorage.GoogleDriveKeyFile]
	GoogleDriveFolderId string
	// Local folder
	FilesDir      string
	AdminPassword string
	AdminMail     string
}

// NewCompta uses env variables to build Compta configuration:
// COMPTA_FILES_DIR, COMPTA_ADMIN_PASSWORD, COMPTA_ADMIN_MAIL
// Google Drive settings uses the optionnal COMPTA_GOOGLE_DRIVE_KEY_FILE, COMPTA_GOOGLE_DRIVE_FOLDER_ID
func NewCompta(checkComptaDrive bool) (Compta, error) {
	keyFile := os.Getenv("COMPTA_GOOGLE_DRIVE_KEY_FILE")
	if keyFile == "" && checkComptaDrive {
		return Compta{}, errors.New("missing env COMPTA_GOOGLE_DRIVE_KEY_FILE")
	}

	folderID := os.Getenv("COMPTA_GOOGLE_DRIVE_FOLDER_ID")
	if folderID == "" && checkComptaDrive {
		return Compta{}, errors.New("missing env COMPTA_GOOGLE_DRIVE_FOLDER_ID")
	}

	filesDir := os.Getenv("COMPTA_FILES_DIR")
	if filesDir == "" {
		return Compta{}, errors.New("missing env COMPTA_FILES_DIR")
	}

	adminPassword := os.Getenv("COMPTA_ADMIN_PASSWORD")
	if adminPassword == "" {
		return Compta{}, errors.New("missing env COMPTA_ADMIN_PASSWORD")
	}

	adminMail := os.Getenv("COMPTA_ADMIN_MAIL")
	if adminMail == "" {
		return Compta{}, errors.New("missing env COMPTA_ADMIN_MAIL")
	}

	filesDir, err := checkDir(filesDir)
	if err != nil {
		return Compta{}, err
	}

	return Compta{
		GoogleDriveKeyFile:  keyFile,
		GoogleDriveFolderId: folderID,
		FilesDir:            filesDir,
		AdminPassword:       adminPassword,
		AdminMail:           adminMail,
	}, nil
}
