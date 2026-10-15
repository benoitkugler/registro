package dons

import (
	"encoding/json"
	"fmt"
	"io"

	"registro/helloasso"
	"registro/sql/dons"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

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

// It returns [false] for other notifications.
func HandleDonNotification(api helloasso.Api, payload io.Reader) (DonDonateur, bool, error) {
	notif, err := parseDonNotification(payload)
	if err != nil {
		return DonDonateur{}, false, err
	}
	if notif.State != "Authorized" {
		return DonDonateur{}, false, nil
	}

	payment, err := api.LoadPayment(notif.Id)
	if err != nil {
		return DonDonateur{}, false, err
	}
	out, err := newDonFromPayment(payment)
	if err != nil {
		return DonDonateur{}, false, err
	}
	return out, true, err
}

type DonDonateur struct {
	Don      dons.Don
	Donateur pr.Identite
}

func newDonFromPayment(payment helloasso.Payment) (DonDonateur, error) {
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
