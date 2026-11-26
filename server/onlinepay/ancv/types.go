// Package ancv implements the direct (merchant) integration with the ANCV
// Chèque-Vacances Connect payment API.
//
// All payload types and algorithms are taken from the ANCV integration kit
// v1.08 (referenced below as "Kit", with the printed page numbers).
// Only the payment-transactions API is implemented (Kit §4.3.3, p. 14); the
// pre-transaction API (Kit §4.3.4, p. 14) is deliberately not used.
package ancv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	// CurrencyEUR is the only currency supported by the ANCV network.
	// Kit §6.2.2.15 (p. 60) : « 978 / EUR / Euros ».
	CurrencyEUR = "978"

	// captureModeNormal makes the validation implicit, so no /execute call
	// is needed. Kit §6.2.2.16 (p. 61) : « Validation automatique : NORMAL ».
	captureModeNormal = "NORMAL"

	// tspdModeNoAdjustment forbids the beneficiary from changing the amount.
	// Kit §6.2.2.16 (p. 61) : « 002 : Chèque Vacances Dématérialisé uniquement,
	// montant non ajustable par le bénéficiaire ».
	tspdModeNoAdjustment = "002"
)

// AcceptedStates are the transaction states for which the payment can be
// recorded. Kit §6.2.2.5 (pp. 56-57) : AUTHORIZED (« la transaction est
// autorisée »), VALIDATED, CONSIGNED, PAID (« Le paiement vers le commerçant
// de la partie payée [...] a été réalisé par l'ANCV »).
const (
	StateAuthorized = "AUTHORIZED"
	StateValidated  = "VALIDATED"
	StateConsigned  = "CONSIGNED"
	StatePaid       = "PAID"
)

// IsAuthorizedState reports whether a transaction state means that the
// beneficiary has authorized the payment. Kit §5.1.2.2 (p. 45) : seul
// AUTHORIZED déclenche le « retour valide ». Les états postérieurs (VALIDATED,
// CONSIGNED, PAID) sont également acceptés pour éviter de manquer un paiement
// si le webhook est reçu tardivement.
func IsAuthorizedState(state string) bool {
	switch state {
	case StateAuthorized, StateValidated, StateConsigned, StatePaid:
		return true
	default:
		return false
	}
}

// merchant identifies the shop calling the API.
// Kit §6.2.2.14 (p. 59) : « shopId : Identifiant technique unique de la
// boutique [...] transmis par l'ANCV ». serviceProviderId is intentionally
// omitted : Kit §6.2.2.14 (p. 59) : « Les commerçants souhaitant se raccorder
// directement au serveur d'acquisition ne doivent pas renseigner ce champ ».
type merchant struct {
	ShopID int64 `json:"shopId"`
}

// Amount is a monetary amount in the currency's minor unit.
// Kit §6.2.2.15 (p. 60) : « total : La valeur totale (TTC) du montant en
// centimes de la devise [...] Exemple : 800 (correspond à 8 €) ».
type Amount struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

// Order describes the merchant's order.
// Kit §6.2.2.15 (p. 60) : « id [...] 64 caractères au maximum », « paymentId
// [...] 40 caractères au maximum », et §4.5.1.3 (p. 17) : « Le couple de
// données (id, paymentId) doit être unique pour un même shopId sur une
// journée ».
type Order struct {
	ID        string `json:"id"`
	PaymentID string `json:"paymentId"`
	Label     string `json:"label,omitempty"`
	Amount    Amount `json:"amount"`
}

// paymentMethod selects the acquisition mode.
// Kit §6.2.2.16 (p. 61) : « tspdMode [...] 3 chiffres fixes » et « captureMode
// [...] NORMAL / DEFERRED ».
type paymentMethod struct {
	CaptureMode string `json:"captureMode"`
	TspdMode    string `json:"tspdMode"`
}

// redirectURLs are called by ANCV on authorization and cancellation.
// Kit §6.2.2.17 (pp. 61-62) : « returnUrl [...] utilisée par le serveur de la
// plateforme dématérialisation pour notifier au commerçant [...] l'autorisation ».
// Kit §5.1.4 (p. 46) : « Que l'IP de provenance de l'appel du webhook ne soit
// pas filtrée [et] Que l'url fournie soit en HTTS [sic] ».
type redirectURLs struct {
	ReturnURL string `json:"returnUrl,omitempty"`
	CancelURL string `json:"cancelUrl,omitempty"`
}

// ApplicationContext is merchant data returned unchanged by ANCV.
// Kit §6.2.5.1 (p. 69) : « Contexte interne de la commande du commerçant à lui
// retourner [...] String de 512 caractères maximum ».
type ApplicationContext struct {
	ReturnContext string `json:"returnContext,omitempty"`
	CustomerID    string `json:"customerId,omitempty"`
}

// InitRequest is the body of POST /payment-transactions.
// Kit §4.5.1.3 (p. 17) : « merchant, order, paymentMethod, redirectUrls,
// applicationContext, requestDate ».
type InitRequest struct {
	Merchant           merchant           `json:"merchant"`
	Order              Order              `json:"order"`
	PaymentMethod      paymentMethod      `json:"paymentMethod"`
	RedirectURLs       redirectURLs       `json:"redirectUrls,omitempty"`
	ApplicationContext ApplicationContext `json:"applicationContext,omitempty"`
	RequestDate        string             `json:"requestDate,omitempty"`
}

// PayRequest is the body of POST /payment-transactions/{id}/payer.
// Kit §4.5.2.3 (p. 21) : « payer [...] Correspond au porteur (bénéficiaire) de
// titres dématérialisés. Si le montant associé au payeur n'est pas transmis, la totalité
// du montant de la transaction lui est attribuée. »
type PayRequest struct {
	Payer       payer  `json:"payer"`
	RequestDate string `json:"requestDate,omitempty"`
}

// payer carries the beneficiary identifier.
// Kit §6.2.3.1 (p. 63) : « L'adresse e-mail du bénéficiaire rattachée à son
// compte CV Connect [...] Le numéro de compte CV Connect (composé de 11 chiffres
// dont clé de Luhn) ».
type payer struct {
	BeneficiaryID string `json:"beneficiaryId"`
	Amount        Amount `json:"amount"`
}

// NewPayRequest builds the body of POST /payment-transactions/{id}/payer.
// Kit §4.5.2.3 (p. 21) : « payer [...] Correspond au porteur (bénéficiaire) de
// titres dématérialisés ». Kit §6.2.2.15 (p. 60) : le montant est en centimes
// et la devise acceptée est « 978 / EUR / Euros ».
func NewPayRequest(beneficiaryID string, montantAsCent int) PayRequest {
	return PayRequest{
		Payer: payer{
			BeneficiaryID: beneficiaryID,
			Amount:        Amount{Total: montantAsCent, Currency: CurrencyEUR},
		},
	}
}

// jsonString decodes a value encoded either as a JSON string or as a JSON
// number. Kit §6.2.3.1 (p. 63) declares beneficiaryId as a String, but the
// example of §4.5.2.6 (p. 22) uses the numeric literal 1536923388807.
type jsonString string

func (s *jsonString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*s = ""
		return nil
	}
	if data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*s = jsonString(v)
		return nil
	}
	if _, err := strconv.ParseFloat(string(data), 64); err != nil {
		return fmt.Errorf("invalid beneficiaryId %q: %s", data, err)
	}
	*s = jsonString(data)
	return nil
}

func (s jsonString) String() string { return string(s) }

// Authorization is a payment authorization granted by the beneficiary.
// Kit §6.2.2.19 (pp. 62-63) : « type [...] moyen de paiement utilisé », « amount
// [...] Montant de l'autorisation ». Le moyen de paiement considéré ici est le
// Chèque-Vacances Connect.
type Authorization struct {
	Number         string `json:"number"`
	Type           string `json:"type"`
	Amount         Amount `json:"amount"`
	ValidationDate string `json:"validationDate"`
	Holder         string `json:"holder"`
}

// PayerInfo is a payer as returned by ANCV.
// Kit §6.2.2.18 (p. 62) : « authorizations [...] n'est transmise que lorsque
// toutes les demandes d'autorisations associées à une transaction ont été
// acceptées ».
type PayerInfo struct {
	BeneficiaryID  jsonString      `json:"beneficiaryId"`
	Amount         Amount          `json:"amount"`
	Authorizations []Authorization `json:"authorizations"`
}

// Transaction is the object returned by the payment-transactions API.
// Kit §6.2.2 (pp. 56-63).
type Transaction struct {
	ID             string        `json:"id"`
	CreationDate   string        `json:"creationDate"`
	UpdateDate     string        `json:"updateDate"`
	ExpirationDate string        `json:"expirationDate"`
	State          string        `json:"state"`
	SubState       string        `json:"subState"`
	Merchant       merchant      `json:"merchant"`
	Order          Order         `json:"order"`
	PaymentMethod  paymentMethod `json:"paymentMethod"`
	RedirectURLs   redirectURLs  `json:"redirectUrls"`
	Payers         []PayerInfo   `json:"payers"`
}

// Response is the common envelope of init, pay, status and webhook calls.
// Kit §4.5.1.5 (p. 18) : « transaction, applicationContext, responseDate ».
// Kit §5.1.3 (p. 46) : « transaction [ou] preTransaction, applicationContext,
// responseDate ».
type Response struct {
	Transaction        *Transaction       `json:"transaction"`
	PreTransaction     json.RawMessage    `json:"preTransaction"`
	ApplicationContext ApplicationContext `json:"applicationContext"`
	ResponseDate       string             `json:"responseDate"`
}

// APIError is returned for a rejected call.
// Kit §4.5.1.5 (p. 18) : « errorCode [...] Contient un code associé à la cause
// du rejet », « errorMessage [...] Contient un message lisible par un humain ».
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"errorCode"`
	Message    string `json:"errorMessage"`
}

func (err APIError) Error() string {
	if err.Code == "" {
		return fmt.Sprintf("ANCV : erreur HTTP %d", err.StatusCode)
	}
	return fmt.Sprintf("ANCV [%s] : %s", err.Code, err.Message)
}

// errorMessages translates the rejection codes of Kit §6.3 (pp. 69-70) into
// messages that can be shown to the family. The kit specifies that « Les
// messages d'erreur ne sont disponibles qu'en langue anglaise ».
var errorMessages = map[string]string{
	"BAD_REQUEST":                           "requête ANCV invalide",
	"INVALID_SEAL":                          "erreur de configuration ANCV (sceau invalide)",
	"MERCHANT_NOT_ALLOWED":                  "le commerçant ANCV n'est pas autorisé",
	"OPERATION_TRANSACTION_NOT_ALLOWED":     "cette opération n'est pas autorisée sur la transaction",
	"OPERATION_PRE_TRANSACTION_NOT_ALLOWED": "cette opération n'est pas autorisée sur la pré-transaction",
	"INSUFFICIENT_BALANCE":                  "le solde du compte Chèque-Vacances est insuffisant",
	"NO_TA_TRANSACTION_PENDING":             "aucune autorisation n'est en attente",
	"TRANSACTION_NOT_VALIDATED_ON_TA_SIDE":  "la transaction n'a pas été validée",
	"TRANSACTION_VALIDATED_ON_TA_SIDE":      "la transaction a déjà été validée",
	"TRANSACTION_NOT_FOUND":                 "transaction ANCV introuvable",
	"BENEFICIARY_NOT_FOUND":                 "compte Chèque-Vacances introuvable",
	"POINT_OF_SALE_NOT_FOUND":               "point de vente ANCV inconnu",
	"OTHER_TRANSACTION_PENDING":             "une autre transaction est en cours",
	"INVALID_TRANSACTION_AMOUNT":            "montant de la transaction invalide",
	"INVALID_TRANSACTION_CURRENCY":          "devise de la transaction invalide",
	"INVALID_PAYER_AMOUNT":                  "montant du payeur invalide",
	"INVALID_TSPD_MODE":                     "mode de paiement invalide",
	"TRANSACTION_EXPIRED":                   "la transaction ANCV a expiré",
	"NO_ACTIVE_DEVICE":                      "aucun appareil actif pour valider le paiement",
}
