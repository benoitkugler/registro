package helloasso

import (
	"fmt"
	"reflect"
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
	_, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
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

	metadata := map[string]any{"ref": "slkmsms", "source": "horizon"}
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
	tu.Assert(t, reflect.DeepEqual(metadata, checkout.Metadata))
	tu.Assert(t, len(checkout.Order.Payments) == 0)
	fmt.Println(intent.Id, intent.RedirectUrl)
}

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
