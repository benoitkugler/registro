//go:build ancv_integration

// Recette integration tests. These call the real ANCV recette API and are
// excluded from normal test runs. From server/, run them with:
//
//	go test -tags ancv_integration ./onlinepay/ancv -run TestRecette -count=1 -v
//
// Valid recette credentials must be set in server/env.sh (ANCV_PROD must not be TRUE).
package ancv

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"registro/config"
	tu "registro/utils/testutils"
)

// recetteConfig loads the real recette credentials. Failing when ANCV_PROD
// is TRUE prevents an accidental call to the production API.
func recetteConfig(t *testing.T) config.ANCV {
	t.Helper()
	tu.LoadEnv(t, "../../env.sh")
	cfg, err := config.NewANCV()
	tu.AssertNoErr(t, err)
	if cfg.Prod {
		t.Fatal("ANCV recette tests require ANCV_PROD not set to TRUE")
	}
	return cfg
}

func recetteAPI(t *testing.T) *Api {
	t.Helper()
	api, err := NewApi(recetteConfig(t))
	tu.AssertNoErr(t, err)
	return api
}

func randomHex(t *testing.T, byteCount int) string {
	t.Helper()
	b := make([]byte, byteCount)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// TestRecetteCreateTransactionAndPay exercises the three sealed calls against
// the recette platform. Kit §6.9 (p. 86) has no notification simulator, so the
// beneficiary is not asked to authorize here : full validation requires the
// application described in Kit §6.10 and document DR3.
func TestRecetteCreateTransactionAndPay(t *testing.T) {
	api := recetteAPI(t)

	orderID := "registro-recette-" + time.Now().UTC().Format("20060102T150405") + "-" + randomHex(t, 8)
	const notificationURL = "https://example.org/api/v1/ancv/notification"
	req, err := api.NewInitRequest(orderID, 100, notificationURL, "registro-recette-test")
	tu.AssertNoErr(t, err)

	initResp, err := api.createTransaction(req)
	tu.AssertNoErr(t, err)
	if initResp.Transaction == nil {
		t.Fatal("recette did not return a transaction")
	}
	tx := initResp.Transaction
	if tx.ID == "" {
		t.Fatal("recette did not return a transaction id")
	}
	if tx.Order.ID != orderID || tx.Order.PaymentID != req.Order.PaymentID {
		t.Fatalf("unexpected order echo: %+v", tx.Order)
	}
	if tx.Order.Amount != (Amount{Total: 100, Currency: CurrencyEUR}) {
		t.Fatalf("unexpected amount echo: %+v", tx.Order.Amount)
	}
	t.Logf("created ANCV recette transaction %s", tx.ID)

	// Kit §4.5.2.6 (p. 22) uses 1536923388807 as beneficiaryId in its
	// /payment-transactions/{id}/payer example.
	payResp, err := api.pay(tx.ID, NewPayRequest("1536923388807", 100))
	tu.AssertNoErr(t, err)
	if payResp.Transaction != nil && payResp.Transaction.ID != tx.ID {
		t.Fatalf("payer returned another transaction: %+v", payResp.Transaction)
	}

	status, err := api.GetTransaction(tx.ID)
	tu.AssertNoErr(t, err)
	if status.Transaction == nil {
		t.Fatal("status call did not return a transaction")
	}
	if status.Transaction.ID != tx.ID || status.Transaction.Order.ID != orderID {
		t.Fatalf("unexpected transaction status: %+v", status.Transaction)
	}
	if status.Transaction.State == "" {
		t.Fatal("recette did not return a transaction state")
	}
	t.Logf("ANCV recette transaction %s state: %s", tx.ID, status.Transaction.State)
}

// TestRecetteRejectsInvalidSeal checks the authentication against the recette
// platform. Kit §6.3 (pp. 69-70) defines INVALID_SEAL as a rejection code.
func TestRecetteRejectsInvalidSeal(t *testing.T) {
	cfg := recetteConfig(t)
	cfg.SealKey += "-invalid"
	api, err := NewApi(cfg)
	tu.AssertNoErr(t, err)

	req, err := api.NewInitRequest(
		"registro-recette-invalid-seal", 100,
		"https://example.org/api/v1/ancv/notification", "invalid-seal-test",
	)
	tu.AssertNoErr(t, err)
	_, err = api.createTransaction(req)
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an APIError, got %v", err)
	}
	if apiErr.Code != "INVALID_SEAL" {
		t.Fatalf("expected INVALID_SEAL, got %s (%v)", apiErr.Code, err)
	}
}
