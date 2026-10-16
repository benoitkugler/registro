package helloasso

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"registro/config"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Helloasso {
	tu.LoadEnv(t, "../env.sh")
	creds, err := config.NewHelloasso()
	tu.AssertNoErr(t, err)
	return creds
}

func TestPing(t *testing.T) {
	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
	tu.Assert(t, api.token != "")
}

func TestDate(t *testing.T) {
	var d time.Time
	err := d.UnmarshalJSON([]byte(`"2025-12-23T13:25:40.276Z"`))
	tu.AssertNoErr(t, err)
	tu.Assert(t, d.Year() == 2025 && d.Month() == time.December && d.Day() == 23)
}

func TestCheckouts(t *testing.T) {
	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)

	// reject invalid values
	_, err = api.InitCheckout(InitCheckout{
		TotalAmount:      100,
		InitialAmount:    100,
		ItemName:         "Test",
		BackUrl:          "https://test.fr/back",
		ErrorUrl:         "https://test.fr/error",
		ReturnUrl:        "https://test.fr/return",
		ContainsDonation: false,
		Payer: Payer{
			FirstName: "Invalid 0",
		},
	})
	tu.AssertErr(t, err)

	type medataT struct{ Reference, Source string }
	metadata := medataT{"slkmsms", "horizon"}

	intent, err := api.InitCheckout(InitCheckout{
		TotalAmount:      1142,
		InitialAmount:    1142,
		ItemName:         "Test",
		BackUrl:          "https://test.fr/back",
		ErrorUrl:         "https://test.fr/error",
		ReturnUrl:        "https://test.fr/return",
		ContainsDonation: false,
		Payer: Payer{
			FirstName: "Camille",
			LastName:  "Martin",
			Email:     "camille@example.com",
		},
		Metadata: metadata,
	})
	tu.AssertNoErr(t, err)

	checkout, err := api.LoadCheckout(intent.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, checkout.CheckoutIntent.Id == intent.Id)
	tu.Assert(t, len(checkout.Order.Payments) == 0)

	var m2 medataT
	err = json.Unmarshal(checkout.Metadata, &m2)
	tu.AssertNoErr(t, err)
	tu.Assert(t, metadata == m2)

	fmt.Println(intent.Id, intent.RedirectUrl)
}

const payloadCheckoutACVE = `
	{"data": {"cardExpirationDate": "2027-12-31T00:00:00Z", "order": {"id": 100058, "date": "2026-10-16T14:40:57.1333076+02:00", "formSlug": "default", "formType": "Checkout", "organizationName": "Acve", "organizationSlug": "acve", "organizationType": "Association1901", "organizationIsUnderColucheLaw": false, "formName": "Checkout", "meta": {"createdAt": "2026-10-16T14:40:39.3140721+02:00", "updatedAt": "2026-10-16T14:40:57.21+02:00"}, "isAnonymous": false, "isAmountHidden": false}, "payer": {"email": "camille@example.com", "address": "", "city": "", "zipCode": "", "country": "FRA", "firstName": "Camille", "lastName": "Martin"}, "items": [{"shareAmount": 1142, "shareItemAmount": 1142, "id": 110776, "amount": 1142, "type": "Payment", "state": "Processed", "name": "Test"}], "displayData": {"formDisplayType": "Checkout", "isTapToPay": false}, "cashOutState": "Transfered", "paymentReceiptUrl": "https://www.helloasso-sandbox.com/associations/acve/checkout/paiement-attestation/100058", "id": 71968, "amount": 1142, "date": "2026-10-16T14:40:57.1333076+02:00", "paymentMeans": "Card", "installmentNumber": 1, "state": "Authorized", "meta": {"createdAt": "2026-10-16T14:40:39.3140721+02:00", "updatedAt": "2026-10-16T14:40:57.21488+02:00"}, "refundOperations": []}, "eventType": "Payment", "metadata": {"ref": "slkmsms", "source": "horizon"}}
	`

func TestCheckoutPayment(t *testing.T) {
	// we rely on manual triggering
	// t.Skip("test requires manuel payment")

	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)

	intentId := int32(282090)
	checkout, err := api.LoadCheckout(intentId)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(checkout.Order.Payments) == 1)
	tu.Assert(t, checkout.IsCompleted())
	tu.Assert(t, checkout.Order.Amount.Total == 1142)
}

func TestParseNotification(t *testing.T) {
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
	const payloadOther = `
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
	const payloadCheckout = `
	{
	"data": {
		"order": {
		"id": 22707,
		"date": "2024-07-10T11:50:49.0890389+02:00",
		"formSlug": "default",
		"formType": "Checkout",
		"organizationName": "bordeauxchess",
		"organizationSlug": "bordeauxchess",
		"organizationType": "Association1901",
		"organizationIsUnderColucheLaw": false,
		"formName": "Checkout",
		"meta": {
			"createdAt": "2024-07-10T11:50:34.0898335+02:00",
			"updatedAt": "2024-07-10T11:50:49.0890389+02:00"
		},
		"isAnonymous": false,
		"isAmountHidden": false
		},
		"payer": {
		"email": "john.doe@example.com",
		"country": "FRA",
		"firstName": "John",
		"lastName": "Doe"
		},
		"items": [
		{
			"shareAmount": 1000,
			"shareItemAmount": 1000,
			"id": 22707,
			"amount": 6000,
			"type": "Payment",
			"state": "Processed",
			"name": "Adhésion football"
		}
		],
		"cashOutState": "Transfered",
		"paymentReceiptUrl": "https://www.helloasso-sandbox.com/associations/bordeauxchess/checkout/paiement-attestation/22707",
		"id": 15222,
		"amount": 1000,
		"date": "2024-07-10T11:50:48.8859116+02:00",
		"paymentMeans": "Card",
		"installmentNumber": 1,
		"state": "Authorized",
		"meta": {
		"createdAt": "2024-07-10T11:50:34.0898335+02:00",
		"updatedAt": "2024-07-10T11:50:48.9171637+02:00"
		},
		"refundOperations": []
	},
	"eventType": "Payment",
	"metadata": {
		"id": 75698555
	}
	}`

	notif, ok, err := ParseNotification(strings.NewReader(payloadCheckout))
	tu.AssertNoErr(t, err)
	tu.Assert(t, ok)
	var metadata struct {
		Id int `json:"id,omitempty"`
	}
	err = json.Unmarshal(notif.Metadata, &metadata)
	tu.AssertNoErr(t, err)
	tu.Assert(t, notif.Data.State == "Authorized")
	tu.Assert(t, notif.Data.Id == 15222)
	tu.Assert(t, notif.Data.Order.FormType == "Checkout")
	tu.Assert(t, metadata.Id == 75698555)

	notif, ok, err = ParseNotification(strings.NewReader(payloadDon))
	tu.AssertNoErr(t, err)
	tu.Assert(t, ok)
	tu.Assert(t, notif.Data.State == "Authorized")
	tu.Assert(t, notif.Data.Id == 45689)
	tu.Assert(t, notif.Data.Order.FormType == "Donation")
	tu.Assert(t, notif.Metadata == nil)

	notif, ok, err = ParseNotification(strings.NewReader(payloadOther))
	tu.AssertNoErr(t, err)
	tu.Assert(t, !ok)

	notif, ok, err = ParseNotification(strings.NewReader(payloadCheckoutACVE))
	tu.AssertNoErr(t, err)
	tu.Assert(t, ok)
	tu.Assert(t, notif.Data.State == "Authorized")
	tu.Assert(t, notif.Data.Id == 71968)
	tu.Assert(t, notif.Data.Order.FormType == "Checkout")
	tu.Assert(t, notif.Metadata != nil)
}

func TestCheckoutNotification(t *testing.T) {
	// generated by paying on checkout 282127
	notif, _, err := ParseNotification(strings.NewReader(payloadCheckoutACVE))
	tu.AssertNoErr(t, err)

	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
	payment, err := api.LoadPayment(notif.Data.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, payment.Order.Id == notif.Data.Order.Id)

	order, err := api.LoadOrder(payment.Order.Id)
	tu.AssertNoErr(t, err)

	tu.Assert(t, order.CheckoutIntentId == 282127)
}
