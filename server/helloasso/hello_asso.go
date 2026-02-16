package helloasso

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"registro/config"
	"registro/sql/dons"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type Api struct {
	config config.Helloasso

	token string // see loadAccessToken
}

func NewApi(config config.Helloasso) Api { return Api{config, ""} }

func (api *Api) baseURL() string {
	if api.config.Sandbox {
		return "https://api.helloasso-sandbox.com"
	}
	return "https://api.helloasso.com"
}

// Ping effectue une requête de test et renvoie l'éventuelle erreur.
func (api Api) Ping() error {
	err := api.loadAccessToken()
	return err
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

func (api *Api) getJSON(endpoint string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, api.baseURL()+endpoint, nil)
	if err != nil {
		return fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+api.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s. Détail : %s", resp.Status, b)
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	return err
}

// It returns [false] for other notifications.
func (api *Api) HandleDonNotification(payload io.Reader) (DonDonateur, bool, error) {
	notif, err := parseDonNotification(payload)
	if err != nil {
		return DonDonateur{}, false, err
	}
	if notif.State != "Authorized" {
		return DonDonateur{}, false, nil
	}

	if err := api.loadAccessToken(); err != nil {
		return DonDonateur{}, false, err
	}
	payment, err := api.loadPayment(notif.Id)
	if err != nil {
		return DonDonateur{}, false, err
	}
	out, err := newDonFromPayment(payment)
	if err != nil {
		return DonDonateur{}, false, err
	}
	return out, true, err
}

func (api *Api) loadPayment(id int32) (paymentHelloAsso, error) {
	var out paymentHelloAsso
	err := api.getJSON(fmt.Sprintf("/v5/payments/%d", id), &out)
	if err != nil {
		return paymentHelloAsso{}, err
	}
	return out, nil
}

type donNotification struct {
	Id    int32  `json:"id"`
	State string `json:"state"`
}

func parseDonNotification(payload io.Reader) (donNotification, error) {
	var n struct {
		EventType string          `json:"eventType"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(payload).Decode(&n); err != nil {
		return donNotification{}, err
	}
	if n.EventType != "Payment" {
		return donNotification{}, nil
	}
	var payment donNotification
	if err := json.Unmarshal(n.Data, &payment); err != nil {
		return donNotification{}, err
	}
	return payment, nil
}

type paymentHelloAsso struct {
	Order struct {
		FormSlug string `json:"formSlug"`
	} `json:"order"`
	Payer struct {
		FirstName   string    `json:"firstName"`
		LastName    string    `json:"lastName"`
		Address     string    `json:"address"`
		ZipCode     string    `json:"zipCode"`
		City        string    `json:"city"`
		Country     string    `json:"country"`
		Email       string    `json:"email"`
		DateOfBirth time.Time `json:"dateOfBirth"`
	} `json:"payer"`
	Date   time.Time `json:"date"`
	Id     int32     `json:"id"`
	Amount int       `json:"amount"` // Montant en centimes
	Items  []struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
	} `json:"items"`
	State string `json:"state"`
}

type DonDonateur struct {
	Don      dons.Don
	Donateur pr.Identite
}

func newDonFromPayment(payment paymentHelloAsso) (DonDonateur, error) {
	don := dons.Don{
		// Hello asso is a french service that only support Euros
		Montant:      dossiers.Montant{Cent: payment.Amount, Currency: dossiers.Euros},
		Date:         shared.NewDateFrom(payment.Date),
		ModePaiement: dossiers.Helloasso,
		Affectation:  payment.Order.FormSlug,
		Details:      fmt.Sprintf("%d", payment.Id), // pourra être modifié
	}
	donateur := pr.Identite{
		Prenom:        payment.Payer.FirstName,
		Nom:           payment.Payer.LastName,
		Adresse:       payment.Payer.Address,
		CodePostal:    payment.Payer.ZipCode,
		Ville:         payment.Payer.City,
		Pays:          newPays(payment.Payer.Country),
		Mail:          payment.Payer.Email,
		DateNaissance: shared.NewDateFrom(payment.Payer.DateOfBirth),
	}
	return DonDonateur{Don: don, Donateur: donateur}, nil
}
