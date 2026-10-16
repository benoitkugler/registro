package horizon

import (
	"registro/config"
	"registro/helloasso"
	tu "registro/utils/testutils"
	"testing"
)

func devCreds(t *testing.T) (config.Helloasso, config.Horizon) {
	tu.LoadEnv(t, "../../env.sh")
	ha, err := config.NewHelloasso()
	tu.AssertNoErr(t, err)

	ho, err := config.NewHorizon()
	tu.AssertNoErr(t, err)

	return ha, ho
}

func TestCheckout(t *testing.T) {
	ha, ho := devCreds(t)
	ct := Controller{ha, ho}

	_, err := ct.createCheckout(createCheckoutIn{})
	tu.AssertErr(t, err)
	_, err = ct.createCheckout(createCheckoutIn{AmountCents: 1_000_000_000})
	tu.AssertErr(t, err)
	_, err = ct.createCheckout(createCheckoutIn{AmountCents: 1000, ItemName: "xxx"})
	tu.AssertErr(t, err)

	intent, err := ct.createCheckout(createCheckoutIn{Reference: "1234-1234-1234-1234", AmountCents: 1000, ItemName: "Horizon 27"})
	tu.AssertNoErr(t, err)
	tu.Assert(t, intent.CheckoutIntentId != 0 && intent.RedirectUrl != "")

	intent, err = ct.createCheckout(createCheckoutIn{Reference: "1234-1234-1234-1234", AmountCents: 1000, ItemName: "Horizon 27",
		Payer: helloasso.Payer{FirstName: "1234"},
	})
	tu.AssertNoErr(t, err) // we have retried without name
	tu.Assert(t, intent.CheckoutIntentId != 0 && intent.RedirectUrl != "")

	api, err := helloasso.NewApi(ha)
	tu.AssertNoErr(t, err)
	checkout, ok, err := loadCheckout(api, intent.CheckoutIntentId)
	tu.AssertNoErr(t, err)
	tu.Assert(t, ok)
	tu.Assert(t, checkout.Reference == "1234-1234-1234-1234" && !checkout.Paid)
}
