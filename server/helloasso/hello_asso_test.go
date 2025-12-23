package helloasso

import (
	"strings"
	"testing"
	"time"

	"registro/config"
	ds "registro/sql/dossiers"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Helloasso {
	tu.LoadEnv(t, "../env.sh")
	creds, err := config.NewHelloasso()
	tu.AssertNoErr(t, err)
	return creds
}

func TestPing(t *testing.T) {
	err := NewApi(devCreds(t)).Ping()
	tu.AssertNoErr(t, err)
}

func TestDate(t *testing.T) {
	var d time.Time
	err := d.UnmarshalJSON([]byte(`"2025-12-23T13:25:40.276Z"`))
	tu.AssertNoErr(t, err)
	tu.Assert(t, d.Year() == 2025 && d.Month() == time.December && d.Day() == 23)
}

// notification generated with sandbox account and a "manual" donation
// with test credit card.
const payloadDon = `
	{
	"data": {
		"order": {
		"id": 71057,
		"date": "2025-12-23T15:56:03.4567544+01:00",
		"formSlug": "1",
		"formType": "Donation",
		"organizationName": "Acve",
		"organizationSlug": "acve",
		"organizationType": "Association1901",
		"organizationIsUnderColucheLaw": false,
		"formName": "Faire un don",
		"meta": {
			"createdAt": "2025-12-23T15:55:43.5038531+01:00",
			"updatedAt": "2025-12-23T15:56:03.5892216+01:00"
		},
		"isAnonymous": false,
		"isAmountHidden": false
		},
		"payer": {
		"email": "bench26@gmail.com",
		"address": "A",
		"city": "Paris",
		"zipCode": "75000",
		"country": "FRA",
		"dateOfBirth": "1990-01-01T00:00:00+01:00",
		"firstName": "Ben",
		"lastName": "Kug"
		},
		"items": [
		{
			"shareAmount": 5500,
			"shareItemAmount": 5500,
			"id": 71057,
			"amount": 5500,
			"type": "Donation",
			"state": "Processed"
		}
		],
		"cashOutState": "Transfered",
		"paymentReceiptUrl": "https://www.helloasso-sandbox.com/associations/acve/formulaires/1/paiement-attestation/71057/45689",
		"id": 45689,
		"amount": 5500,
		"date": "2025-12-23T15:56:03.4567544+01:00",
		"paymentMeans": "Card",
		"installmentNumber": 1,
		"state": "Authorized",
		"meta": {
		"createdAt": "2025-12-23T15:55:43.5038531+01:00",
		"updatedAt": "2025-12-23T15:56:03.5243055+01:00"
		},
		"refundOperations": []
	},
	"eventType": "Payment"
	}
	`

func TestNotifications(t *testing.T) {
	dn, err := parseDonNotification(strings.NewReader(payloadDon))
	tu.AssertNoErr(t, err)
	tu.Assert(t, dn.Id == 45689 && dn.State == "Authorized")

	payloadOther1 := `
	{
	"data": {
		"organizationName": "Acve",
		"tiers": [
		{
			"id": 13572,
			"tierType": "Donation",
			"price": 2000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13573,
			"tierType": "Donation",
			"price": 5000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": true
		},
		{
			"id": 13574,
			"tierType": "Donation",
			"price": 10000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13575,
			"tierType": "Donation",
			"price": 15000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13576,
			"tierType": "Donation",
			"vatRate": 0,
			"minAmount": 50,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13577,
			"tierType": "MonthlyDonation",
			"price": 1000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13578,
			"tierType": "MonthlyDonation",
			"price": 2000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": true
		},
		{
			"id": 13579,
			"tierType": "MonthlyDonation",
			"price": 3000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13580,
			"tierType": "MonthlyDonation",
			"price": 4000,
			"vatRate": 0,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		},
		{
			"id": 13581,
			"tierType": "MonthlyDonation",
			"vatRate": 0,
			"minAmount": 50,
			"paymentFrequency": "Single",
			"isEligibleTaxReceipt": true,
			"isFavorite": false
		}
		],
		"activityTypeId": 0,
		"currency": "EUR",
		"meta": {
		"createdAt": "2025-12-23T15:52:17.7732682+01:00",
		"updatedAt": "2025-12-23T15:52:17.7734869+01:00"
		},
		"state": "Draft",
		"title": "Faire un don",
		"privateTitle": "don",
		"widgetButtonUrl": "https://www.helloasso-sandbox.com/associations/acve/formulaires/1/widget-bouton",
		"widgetFullUrl": "https://www.helloasso-sandbox.com/associations/acve/formulaires/1/widget",
		"formSlug": "1",
		"formType": "Donation",
		"url": "https://www.helloasso-sandbox.com/associations/acve/formulaires/1",
		"organizationSlug": "acve"
	},
	"eventType": "Form"
	}
	`
	bn, err := parseDonNotification(strings.NewReader(payloadOther1))
	tu.AssertNoErr(t, err)
	tu.Assert(t, bn.State == "")
}

func TestHandleDon(t *testing.T) {
	api := NewApi(devCreds(t))
	don, ok, err := api.HandleDonNotification(strings.NewReader(payloadDon))
	tu.AssertNoErr(t, err)
	tu.Assert(t, ok)
	tu.Assert(t, don.Don.Montant == ds.NewEuros(55))
	tu.Assert(t, don.Donateur.DateNaissance == shared.NewDate(1990, time.January, 1))
}
