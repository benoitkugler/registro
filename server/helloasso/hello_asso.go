package helloasso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"registro/config"
)

type Api struct {
	config config.Helloasso

	token string // see loadAccessToken
}

func NewApi(config config.Helloasso) (Api, error) {
	api := Api{config, ""}
	err := api.loadAccessToken()
	return api, err
}

func (api *Api) baseURL() string {
	if api.config.Sandbox {
		return "https://api.helloasso-sandbox.com"
	}
	return "https://api.helloasso.com"
}

func (api *Api) loadAccessToken() error {
	params := url.Values{}
	params.Set("client_id", api.config.ID)
	params.Set("client_secret", api.config.Secret)
	params.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, api.baseURL()+"/oauth2/token", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		var errPayload struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&errPayload); err != nil {
			return fmt.Errorf("La réponse d'HelloAsso est invalide : %s", err)
		}
		return fmt.Errorf("Erreur renvoyée par HelloAsso : %s", errPayload.ErrorDescription)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("internal error : invalid HelloAsso response status %s", resp.Status)
	}

	var payload struct {
		AccessToken  string `json:"access_token"`  //	The JWT token to use in future requests
		RefreshToken string `json:"refresh_token"` //	Token used to refresh the token and get a new JWT token after expiration
		TokenType    string `json:"token_type"`    //	Token Type : always "bearer"
		ExpiresIn    int    `json:"expires_in"`    //	The lifetime in seconds of the access token
	}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return fmt.Errorf("La réponse d'HelloAsso est invalide : %s", err)
	}

	api.token = payload.AccessToken

	return nil
}

func (api *Api) getJSON(endpoint string, out any) error {
	fullURL := api.baseURL() + endpoint
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+api.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête GET vers HelloAsso (%s) a échoué : %s", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("La requête GET vers HelloAsso (%s) a échoué : %s. Détail : %s", fullURL, resp.Status, b)
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	return err
}

func (api *Api) LoadPayment(id int32) (Payment, error) {
	var out Payment
	err := api.getJSON(fmt.Sprintf("/v5/payments/%d", id), &out)
	if err != nil {
		return Payment{}, err
	}
	return out, nil
}

type Payer struct {
	FirstName   string    `json:"firstName,omitempty"`
	LastName    string    `json:"lastName,omitempty"`
	Email       string    `json:"email,omitempty"`
	DateOfBirth time.Time `json:"dateOfBirth,omitzero"`
	Address     string    `json:"address,omitempty"`
	City        string    `json:"city,omitempty"`
	ZipCode     string    `json:"zipCode,omitempty"`
	Country     string    `json:"country,omitempty"`
}

type Payment struct {
	Order struct {
		Id       int32  `json:"id"`
		FormSlug string `json:"formSlug"`
	} `json:"order"`
	Payer  Payer     `json:"payer"`
	Date   time.Time `json:"date"`
	Id     int32     `json:"id"`
	Amount int       `json:"amount"` // Montant en centimes
	Items  []struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
	} `json:"items"`
	State string `json:"state"`
}

// send and receive JSON
func (api *Api) post(endpoint string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	fullURL := api.baseURL() + endpoint
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+api.token)
	req.Header.Add("content-type", "application/json-patch+json")
	req.Header.Add("accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête POST vers HelloAsso (%s) a échoué : %s", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("La requête POST vers HelloAsso (%s) a échoué : %s. Détail : %s", fullURL, resp.Status, b)
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	return err
}

type InitCheckout struct {
	TotalAmount      int32  `json:"totalAmount"`
	InitialAmount    int32  `json:"initialAmount"`
	ItemName         string `json:"itemName"`
	BackUrl          string `json:"backUrl"`
	ErrorUrl         string `json:"errorUrl"`
	ReturnUrl        string `json:"returnUrl"`
	ContainsDonation bool   `json:"containsDonation"`
	Payer            Payer  `json:"payer"`
	Metadata         any    `json:"metadata"`
}

type CheckoutIntent struct {
	Id          int32  `json:"id"`          // Id of the checkout intent
	RedirectUrl string `json:"redirectUrl"` // Url where the contributor must be redirected to
}

func (api *Api) InitCheckout(args InitCheckout) (CheckoutIntent, error) {
	var response CheckoutIntent
	err := api.post(fmt.Sprintf("/v5/organizations/%s/checkout-intents", api.config.Slug), args, &response)
	return response, err
}

type Checkout struct {
	CheckoutIntent
	Order    Order           `json:"order"`
	Metadata json.RawMessage `json:"metadata"`
}

// IsCompleted returns true if at least one payment is Authorized
func (ch Checkout) IsCompleted() bool {
	for _, payment := range ch.Order.Payments {
		if payment.State == "Authorized" {
			return true
		}
	}
	return false
}

type Order struct {
	Id     int32 `json:"id"`
	Amount struct {
		Total    int32 `json:"total"`    // Total amount in cents
		Vat      int32 `json:"vat"`      // Vat amount in cents
		Discount int32 `json:"discount"` // Discount amount in cents
	} `json:"amount"`
	Payments []struct {
		Id     int32  `json:"id"`
		Amount int    `json:"amount"` // Montant en centimes
		State  string `json:"state"`
	} `json:"payments"`
	CheckoutIntentId int32 `json:"checkoutIntentId,omitempty"`
}

func (api *Api) LoadCheckout(checkoutIntentId int32) (Checkout, error) {
	var out Checkout
	err := api.getJSON(fmt.Sprintf("/v5/organizations/%s/checkout-intents/%d", api.config.Slug, checkoutIntentId), &out)
	return out, err
}

func (api *Api) LoadOrder(orderId int32) (Order, error) {
	var out Order
	err := api.getJSON(fmt.Sprintf("/v5/orders/%d", orderId), &out)
	return out, err
}

// PaymentNotification is send by Helloasso on the global webhook endpoint registered
// for the association. We only care for donations and checkouts.
type PaymentNotification struct {
	Data struct {
		Id    int32  `json:"id"` // Payment id
		State string `json:"state"`
		Order struct {
			Id       int32  `json:"id"`
			FormType string `json:"formType"` // "Donation" or "Checkout"
		} `json:"order"`
	}
	Metadata json.RawMessage // to be decoded based on the notification type
}

// ParseNotification parse any HelloAsso notification, but returns false for notifications
// other than "Payment".
func ParseNotification(payload io.Reader) (PaymentNotification, bool, error) {
	var n struct {
		EventType string          `json:"eventType"`
		Data      json.RawMessage `json:"data"`
		Metadata  json.RawMessage `json:"metadata"`
	}
	if err := json.NewDecoder(payload).Decode(&n); err != nil {
		return PaymentNotification{}, false, err
	}
	if n.EventType != "Payment" {
		return PaymentNotification{}, false, nil
	}
	payment := PaymentNotification{Metadata: n.Metadata}
	if err := json.Unmarshal(n.Data, &payment.Data); err != nil {
		return PaymentNotification{}, true, err
	}
	return payment, true, nil
}
