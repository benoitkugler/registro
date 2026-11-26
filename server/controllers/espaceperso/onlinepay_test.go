package espaceperso

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"registro/onlinepay"
	ds "registro/sql/dossiers"
	tu "registro/utils/testutils"

	"github.com/labstack/echo/v4"
)

func TestPaiementOptions(t *testing.T) {
	taux := ds.Taux{Euros: 1000, FrancsSuisse: 1106}

	// Small amounts: the minimum is clamped to the amount asked.
	options := paiementOptions(taux, ds.NewEuros(30))
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	for _, option := range options {
		if option.Min != option.Max {
			t.Errorf("min and max should match for a small amount: %+v", option)
		}
	}

	// Larger amounts: the minimum is 50 units in each currency.
	options = paiementOptions(taux, ds.NewEuros(100))
	for _, option := range options {
		if option.Min != minOnlinePaymentCent {
			t.Errorf("unexpected minimum: %+v", option)
		}
		if option.Max <= option.Min {
			t.Errorf("maximum should be greater than minimum: %+v", option)
		}
	}
}

type ignoringProvider struct{}

func (ignoringProvider) InitGateway(onlinepay.GatewayStart) (string, error) {
	return "", nil
}

func (ignoringProvider) ReceivePayment(ds.DB, io.ReadCloser, http.Header) (ds.Paiement, bool, error) {
	return ds.Paiement{}, false, nil
}

func TestWebhookIgnoresOtherNotifications(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/onlinepay/webhook", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ct := Controller{provider: ignoringProvider{}}
	tu.AssertNoErr(t, ct.HandleOnlinePaymentNotification(c))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
