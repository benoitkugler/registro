package onlinepay

import (
	"testing"

	"registro/config"
	tu "registro/utils/testutils"
)

func TestNewProvider(t *testing.T) {
	t.Setenv("STRIPE_KEY", "test-key")
	t.Setenv("STRIPE_WEBHOOK", "test-webhook")
	t.Setenv("PAYREXX_INSTANCE", "test-instance")
	t.Setenv("PAYREXX_API_KEY", "test-key")

	provider, err := NewProvider(config.OnlinePayStripe)
	tu.AssertNoErr(t, err)
	if provider == nil {
		t.Fatal("expected a Stripe provider")
	}

	provider, err = NewProvider(config.OnlinePayPayrexx)
	tu.AssertNoErr(t, err)
	if provider == nil {
		t.Fatal("expected a Payrexx provider")
	}

	if _, err := NewProvider("unknown"); err == nil {
		t.Fatal("expected an error for an unknown provider")
	}
}
