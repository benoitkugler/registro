package ancv

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"registro/config"
)

const (
	// Kit §4.2.1 (p. 13) : « https://recette.connect.ancv.com/acquisition/api/public/v1 ».
	recetteURL = "https://recette.connect.ancv.com/acquisition/api/public/v1"
	// Kit §4.2.2 (p. 13) : « https://connect.ancv.com/acquisition/api/public/v1 ».
	productionURL = "https://connect.ancv.com/acquisition/api/public/v1"
)

// apiTimeout bounds every ANCV call. Without it, http.DefaultClient has no
// timeout and a stalled ANCV server could block a family or webhook request.
const apiTimeout = 10 * time.Second

// Api calls the payment-transactions API of ANCV.
// Kit §4.3.3 (p. 14) : « L'accès à chaque opération est sécurisé et inclut une
// vérification de la validité du Commerçant et de l'Intermédiaire de Service ».
type Api struct {
	config  config.ANCV
	baseURL *url.URL
}

// NewApi builds a client from the environment configuration.
func NewApi(cfg config.ANCV) (*Api, error) {
	base := recetteURL
	if cfg.Prod {
		base = productionURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	return &Api{baseURL: u, config: cfg}, nil
}

// NewInitRequest builds the body of POST /payment-transactions from the
// application-level data. ANCV-specific settings, including the paymentId,
// stay in this package.
//
// Kit §4.5.1.3 (p. 17) : « merchant, order, paymentMethod, redirectUrls,
// applicationContext, requestDate ». Kit §6.2.2.16 (p. 61) prescribes
// « Validation automatique : NORMAL » and « 002 : Chèque Vacances
// Dématérialisé uniquement, montant non ajustable par le bénéficiaire ».
// Kit §6.2.5.1 (p. 69) : « Contexte interne de la commande du commerçant à lui
// retourner [...] String de 512 caractères maximum ».
func (c *Api) NewInitRequest(orderID string, montantAsCent int, notificationURL, returnContext string) (InitRequest, error) {
	paymentID, err := newPaymentID()
	if err != nil {
		return InitRequest{}, err
	}
	return InitRequest{
		Merchant: merchant{ShopID: c.config.ShopID},
		Order: Order{
			ID:        orderID,
			PaymentID: paymentID,
			Amount:    Amount{Total: montantAsCent, Currency: CurrencyEUR},
		},
		PaymentMethod: paymentMethod{
			CaptureMode: captureModeNormal,
			TspdMode:    tspdModeNoAdjustment,
		},
		RedirectURLs: redirectURLs{
			ReturnURL: notificationURL,
			CancelURL: notificationURL,
		},
		ApplicationContext: ApplicationContext{ReturnContext: returnContext},
	}, nil
}

// newPaymentID returns a unique identifier for the (id, paymentId) pair.
// Kit §4.5.1.3 (p. 17) : « Le couple de données (id, paymentId) doit être
// unique pour un même shopId sur une journée » ; paymentId is limited to 40
// characters (§6.2.2.15, p. 60). 16 random bytes give 32 hex characters.
func newPaymentID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func formatShopID(id int64) string { return strconv.FormatInt(id, 10) }
func formatAmount(cent int) string { return strconv.Itoa(cent) }

// PaymentResult is the outcome of initializing a transaction and paying it.
type PaymentResult struct {
	TransactionID string
	State         string
}

// CreateAndPay initializes a transaction and submits it to the beneficiary.
// Kit §4.5.1 (p. 17) : « La première étape du paiement [...] est
// l'initialisation d'une transaction de paiement ». Kit §4.5.2 (p. 20) : « La
// mise en paiement consiste à soumettre une transaction [...] pour acquisition
// du paiement par un bénéficiaire ».
func (c *Api) CreateAndPay(req InitRequest, beneficiaryID string) (PaymentResult, error) {
	if !validANCVBeficiary(beneficiaryID) {
		return PaymentResult{}, errors.New("Identifiant Chèque-Vacances invalide (numéro de compte ou adresse mail).")
	}
	initResp, err := c.createTransaction(req)
	if err != nil {
		return PaymentResult{}, err
	}
	if initResp.Transaction == nil || initResp.Transaction.ID == "" {
		return PaymentResult{}, errors.New("réponse ANCV incomplète lors de l'initialisation.")
	}
	payResp, err := c.pay(initResp.Transaction.ID, NewPayRequest(beneficiaryID, req.Order.Amount.Total))
	if err != nil {
		return PaymentResult{}, err
	}
	result := PaymentResult{TransactionID: initResp.Transaction.ID}
	if payResp.Transaction != nil {
		result.State = payResp.Transaction.State
	}
	return result, nil
}

// validANCVBeficiary checks the two accepted forms of beneficiaryId.
// Kit §6.2.3.1 (p. 63) : « L'adresse e-mail du bénéficiaire rattachée à son
// compte CV Connect [...] Le numéro de compte CV Connect (composé de 11 chiffres
// dont clé de Luhn) ».
func validANCVBeficiary(id string) bool {
	if id == "" || len(id) > 254 {
		return false
	}
	// La doc annonce 11 chiffres, mais les exemples de l'API utilisent un
	// numéro à 13 chiffres (§4.5.2.6, p. 22). On accepte les deux formats.
	if len(id) == 11 || len(id) == 13 {
		allDigits := true
		for _, r := range id {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return luhnValid(id)
		}
	}
	address, err := mail.ParseAddress(id)
	return err == nil && address.Address == id
}

func luhnValid(id string) bool {
	sum, alternate := 0, false
	for i := len(id) - 1; i >= 0; i-- {
		digit := int(id[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}

// createTransaction initializes a payment transaction.
// Kit §4.5.1 (p. 17) : « La première étape du paiement [...] est
// l'initialisation d'une transaction de paiement » (POST /payment-transactions).
func (c *Api) createTransaction(req InitRequest) (Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	header := securityHeader(c.config.SealVersion, seal(c.config.SealKey, initSealFields(req.Merchant, req.Order)...))
	var out Response
	if err := c.do(http.MethodPost, "/payment-transactions", body, header, &out); err != nil {
		return Response{}, err
	}
	return out, nil
}

// pay submits the transaction to the beneficiary for authorization.
// Kit §4.5.2 (p. 20) : « La mise en paiement consiste à soumettre une
// transaction [...] pour acquisition du paiement par un bénéficiaire sur un
// montant en titre dématérialisé » (POST /payment-transactions/{id}/payer).
//
// Kit §4.5.2 (p. 21) : « En cas de réception d'une erreur technique (avec un
// statut HTTP 500 par exemple), la bonne procédure est de réaliser un appel à
// l'API de Demande d'état de transaction pour vérifier si la Mise en paiement
// a été réalisée ou non. Dans le cas contraire, il est possible que le
// bénéficiaire aille au bout de la procédure de paiement sans que l'IS soit
// informé [...] que la transaction a bien été autorisée ». On ré-interroge donc
// toujours l'état après une erreur technique.
func (c *Api) pay(transactionID string, req PayRequest) (Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	header := securityHeader(c.config.SealVersion, seal(c.config.SealKey, paySealFields(transactionID, req.Payer)...))
	path := "/payment-transactions/" + url.PathEscape(transactionID) + "/payer"
	var out Response
	err = c.do(http.MethodPost, path, body, header, &out)
	if err == nil {
		return out, nil
	}

	var apiErr APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode < 500 {
		return Response{}, err // rejet fonctionnel, pas de rattrapage
	}

	latest, checkErr := c.GetTransaction(transactionID)
	if checkErr == nil && latest.Transaction != nil && IsAuthorizedState(latest.Transaction.State) {
		return latest, nil
	}
	if checkErr != nil {
		return Response{}, fmt.Errorf("%w (vérification d'état impossible : %s)", err, checkErr)
	}
	state := "inconnu"
	if latest.Transaction != nil {
		state = latest.Transaction.State
	}
	return Response{}, fmt.Errorf("%w (état après vérification : %s)", err, state)
}

// GetTransaction returns the authoritative state of a transaction.
// Kit §4.5.3 (p. 23) : « En cas de doute sur l'état de la transaction [...],
// il appartient à l'Intermédiaire de Service de réaliser un appel à la demande
// d'état de transaction » (GET /payment-transactions/{id}).
func (c *Api) GetTransaction(transactionID string) (Response, error) {
	header := securityHeader(c.config.SealVersion, seal(c.config.SealKey, statusSealFields(transactionID)...))
	path := "/payment-transactions/" + url.PathEscape(transactionID)
	var out Response
	if err := c.do(http.MethodGet, path, nil, header, &out); err != nil {
		return Response{}, err
	}
	return out, nil
}

// do performs a sealed call and decodes either a success response or an
// APIError. Kit §4.3.1 (p. 13) : « Content-Type : application/json » et
// « Ancv-Security : {algo.version.seal} ».
func (c *Api) do(method, path string, body []byte, header string, out interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	u := *c.baseURL
	u.Path = strings.TrimSuffix(u.Path, "/") + path

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ANCV-Security", header)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("appel ANCV impossible : %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("réponse ANCV illisible : %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr APIError
		if err := json.Unmarshal(data, &apiErr); err != nil || apiErr.Code == "" {
			return APIError{StatusCode: resp.StatusCode}
		}
		apiErr.StatusCode = resp.StatusCode
		apiErr.Message = translateError(apiErr.Code, apiErr.Message)
		return apiErr
	}
	if out != nil && len(data) != 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("réponse ANCV invalide : %w", err)
		}
	}
	return nil
}

func translateError(code, message string) string {
	if translated, has := errorMessages[code]; has {
		return translated
	}
	if message != "" {
		return message
	}
	return code
}
