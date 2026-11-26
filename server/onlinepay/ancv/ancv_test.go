package ancv

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"registro/config"
)

// TestSealVector uses the official example of Kit §6.7 (p. 81) :
// secretKey = "663768ff68ad8ea6768bbf65163e9b0a",
// messageToHash = "10000065&100016&panier-33455&42556&500",
// expected = "mfy6VhbdyiErpfvQ3AvnKwU39W_ae9MfuaVurEg-KjE".
func TestSealVector(t *testing.T) {
	const key = "663768ff68ad8ea6768bbf65163e9b0a"
	const want = "mfy6VhbdyiErpfvQ3AvnKwU39W_ae9MfuaVurEg-KjE"

	// Le vecteur officiel contient serviceProviderId ; notre intégration
	// directe n'envoie pas ce champ (Kit §6.2.2.14, p. 59), mais l'algorithme
	// de scellement est le même.
	if got := seal(key, "10000065", "100016", "panier-33455", "42556", "500"); got != want {
		t.Fatalf("unexpected seal: got %s, want %s", got, want)
	}

	merchant := merchant{ShopID: 10000065}
	order := Order{
		ID:        "panier-33455",
		PaymentID: "42556",
		Amount:    Amount{Total: 500, Currency: CurrencyEUR},
	}
	fields := initSealFields(merchant, order)
	if got, expected := strings.Join(fields, "&"), "10000065&&panier-33455&42556&500"; got != expected {
		t.Fatalf("unexpected sealed fields: %q", got)
	}
	if got := seal(key, fields...); got != seal(key, "10000065", "panier-33455", "42556", "500") {
		t.Fatalf("empty serviceProviderId should be skipped: %s", got)
	}
}

func TestSealSkipsEmptyFields(t *testing.T) {
	const key = "secret"
	if got, want := seal(key, "a", "", "b", ""), seal(key, "a", "b"); got != want {
		t.Fatalf("empty fields must be skipped: %s != %s", got, want)
	}
	if got, want := securityHeader("version-3620", "abcd"), "HmacSHA256.version-3620.abcd"; got != want {
		t.Fatalf("unexpected header: %s", got)
	}
}

func TestNewApiSelectsEnvironment(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.ANCV
		want string
	}{
		{name: "default is recette", cfg: config.ANCV{}, want: recetteURL},
		{name: "prod is opt-in", cfg: config.ANCV{Prod: true}, want: productionURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api, err := NewApi(tc.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if got := api.baseURL.String(); got != tc.want {
				t.Fatalf("unexpected base URL: got %s, want %s", got, tc.want)
			}
		})
	}
}

const testSealKey = "663768ff68ad8ea6768bbf65163e9b0a"

// newClientWithMockANCVServer returns an Api pointed at a local httptest
// server that stands in for the ANCV API. The handler acts as the ANCV server:
// it inspects the client's requests and returns canned responses, so no call
// reaches the real ANCV platform.
func newClientWithMockANCVServer(t *testing.T, mockANCVServer http.HandlerFunc) *Api {
	t.Helper()
	server := httptest.NewServer(mockANCVServer)
	t.Cleanup(server.Close)

	client, err := NewApi(config.ANCV{
		ShopID:      10000065,
		SealKey:     testSealKey,
		SealVersion: "version-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	base, err := url.Parse(server.URL + "/acquisition/api/public/v1")
	if err != nil {
		t.Fatal(err)
	}
	client.baseURL = base
	return client
}

func TestNewInitRequest(t *testing.T) {
	api := &Api{config: config.ANCV{ShopID: 10000065}}
	got, err := api.NewInitRequest(
		"panier-33455", 3500,
		"https://example.org/api/v1/ancv/notification", "dossier-token",
	)
	if err != nil {
		t.Fatal(err)
	}
	want := InitRequest{
		Merchant: merchant{ShopID: 10000065},
		Order: Order{
			ID:     "panier-33455",
			Amount: Amount{Total: 3500, Currency: CurrencyEUR},
		},
		PaymentMethod: paymentMethod{
			CaptureMode: captureModeNormal,
			TspdMode:    tspdModeNoAdjustment,
		},
		RedirectURLs: redirectURLs{
			ReturnURL: "https://example.org/api/v1/ancv/notification",
			CancelURL: "https://example.org/api/v1/ancv/notification",
		},
		ApplicationContext: ApplicationContext{ReturnContext: "dossier-token"},
	}
	paymentID := got.Order.PaymentID
	got.Order.PaymentID = ""
	if got != want {
		t.Fatalf("unexpected init request:\n got: %+v\nwant: %+v", got, want)
	}
	if len(paymentID) != 32 || strings.Trim(paymentID, "0123456789abcdef") != "" {
		t.Fatalf("unexpected generated paymentId: %q", paymentID)
	}
}

func TestNewInitRequestGeneratesDistinctPaymentIDs(t *testing.T) {
	api := &Api{config: config.ANCV{ShopID: 10000065}}
	const notificationURL = "https://example.org/api/v1/ancv/notification"
	first, err := api.NewInitRequest("panier-33455", 3500, notificationURL, "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	second, err := api.NewInitRequest("panier-33455", 3500, notificationURL, "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	if first.Order.PaymentID == second.Order.PaymentID {
		t.Fatalf("paymentId should be unique: %s", first.Order.PaymentID)
	}
}

func TestNewPayRequest(t *testing.T) {
	got := NewPayRequest("1536923388807", 3500)
	want := PayRequest{Payer: payer{
		BeneficiaryID: "1536923388807",
		Amount:        Amount{Total: 3500, Currency: CurrencyEUR},
	}}
	if got != want {
		t.Fatalf("unexpected pay request:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestCreateTransaction(t *testing.T) {
	var gotPath, gotHeader, gotBody string
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("ANCV-Security")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"INITIALIZED"},"responseDate":"2018-08-28T10:18:00Z"}`)
	})

	req := InitRequest{
		Merchant: merchant{ShopID: 10000065},
		Order: Order{
			ID:        "panier-33455",
			PaymentID: "42556",
			Amount:    Amount{Total: 500, Currency: CurrencyEUR},
		},
		PaymentMethod: paymentMethod{CaptureMode: captureModeNormal, TspdMode: tspdModeNoAdjustment},
	}
	resp, err := client.createTransaction(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/acquisition/api/public/v1/payment-transactions"; gotPath != want {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if want := securityHeader("version-test", seal(testSealKey, initSealFields(req.Merchant, req.Order)...)); gotHeader != want {
		t.Fatalf("unexpected security header: %s", gotHeader)
	}
	if !strings.Contains(gotBody, `"shopId":10000065`) || strings.Contains(gotBody, "serviceProviderId") {
		t.Fatalf("unexpected body: %s", gotBody)
	}
	if resp.Transaction == nil || resp.Transaction.ID != "trx123" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateAndPay(t *testing.T) {
	var calls []string
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/acquisition/api/public/v1/payment-transactions":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"INITIALIZED"}}`)
		case "/acquisition/api/public/v1/payment-transactions/trx123/payer":
			body, _ := io.ReadAll(r.Body)
			var payReq PayRequest
			if err := json.Unmarshal(body, &payReq); err != nil {
				t.Fatal(err)
			}
			if want := NewPayRequest("1536923388807", 3500); payReq != want {
				t.Fatalf("unexpected pay request: got %+v, want %+v", payReq, want)
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"PROCESSING"}}`)
		default:
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}
	})

	req, err := client.NewInitRequest("panier-33455", 3500, "https://example.org/api/v1/ancv/notification", "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.CreateAndPay(req, "1536923388807")
	if err != nil {
		t.Fatal(err)
	}
	if want := (PaymentResult{TransactionID: "trx123", State: "PROCESSING"}); got != want {
		t.Fatalf("unexpected result: got %+v, want %+v", got, want)
	}
	wantCalls := []string{
		"POST /acquisition/api/public/v1/payment-transactions",
		"POST /acquisition/api/public/v1/payment-transactions/trx123/payer",
	}
	if strings.Join(calls, "|") != strings.Join(wantCalls, "|") {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestCreateAndPayInitErrorSkipsPayment(t *testing.T) {
	payCalls := 0
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/payer") {
			payCalls++
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"errorCode":"MERCHANT_NOT_ALLOWED","errorMessage":"not allowed"}`)
	})

	req, err := client.NewInitRequest("panier-33455", 3500, "https://example.org/api/v1/ancv/notification", "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateAndPay(req, "1536923388807")
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an APIError, got %v", err)
	}
	if payCalls != 0 {
		t.Fatalf("payment should be skipped after an init error: %d calls", payCalls)
	}
}

func TestCreateAndPayIncompleteInitResponse(t *testing.T) {
	payCalls := 0
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/payer") {
			payCalls++
		}
		_, _ = io.WriteString(w, `{}`)
	})

	req, err := client.NewInitRequest("panier-33455", 3500, "https://example.org/api/v1/ancv/notification", "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateAndPay(req, "1536923388807"); err == nil {
		t.Fatal("expected an error for an incomplete init response")
	}
	if payCalls != 0 {
		t.Fatalf("payment should be skipped without a transaction id: %d calls", payCalls)
	}
}

func TestCreateAndPayInvalidBeneficiary(t *testing.T) {
	calls := 0
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
	})

	req, err := client.NewInitRequest("panier-33455", 3500, "https://example.org/api/v1/ancv/notification", "dossier-token")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateAndPay(req, "not-a-beneficiary")
	if err == nil || !strings.Contains(err.Error(), "invalide") {
		t.Fatalf("expected a beneficiary validation error, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("no ANCV call expected, got %d", calls)
	}
}

func TestValidANCVBeficiary(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"1536923388807", true},                            // 13 digits, Kit §4.5.2.6
		{"79927398713", true},                              // 11 digits with a valid Luhn key
		{"79927398712", false},                             // invalid Luhn key
		{"jean.dupont@example.org", true},                  // email address
		{"Jean <jean@example.org>", false},                 // display name is not accepted
		{"jean.dupont@", false},                            // malformed email
		{"", false},                                        // empty
		{strings.Repeat("a", 255) + "@example.org", false}, // too long
	}
	for _, tc := range cases {
		if got := validANCVBeficiary(tc.id); got != tc.want {
			t.Errorf("validANCVBeficiary(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestPayAndStatus(t *testing.T) {
	var calls []string
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			var req PayRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatal(err)
			}
			if want := securityHeader("version-test", seal(testSealKey, paySealFields("trx123", req.Payer)...)); r.Header.Get("ANCV-Security") != want {
				t.Fatalf("unexpected seal on payer: %s", r.Header.Get("ANCV-Security"))
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"PROCESSING"}}`)
			return
		}
		if want := securityHeader("version-test", seal(testSealKey, statusSealFields("trx123")...)); r.Header.Get("ANCV-Security") != want {
			t.Fatalf("unexpected seal on status: %s", r.Header.Get("ANCV-Security"))
		}
		_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"AUTHORIZED","payers":[{"beneficiaryId":1536923388807,"authorizations":[{"type":"CVCo","amount":{"total":3500,"currency":"978"}}]}]}}`)
	})

	payer := payer{BeneficiaryID: "1536923388807", Amount: Amount{Total: 3500, Currency: CurrencyEUR}}
	resp, err := client.pay("trx123", PayRequest{Payer: payer})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Transaction.State != "PROCESSING" {
		t.Fatalf("unexpected state: %s", resp.Transaction.State)
	}

	status, err := client.GetTransaction("trx123")
	if err != nil {
		t.Fatal(err)
	}
	if !IsAuthorizedState(status.Transaction.State) {
		t.Fatalf("transaction should be authorized: %s", status.Transaction.State)
	}
	authorizations := status.Transaction.Payers[0].Authorizations
	if len(authorizations) != 1 || authorizations[0].Amount.Total != 3500 {
		t.Fatalf("unexpected authorizations: %+v", authorizations)
	}
	if got := status.Transaction.Payers[0].BeneficiaryID.String(); got != "1536923388807" {
		t.Fatalf("numeric beneficiaryId was not decoded: %s", got)
	}
	if want := []string{"POST /acquisition/api/public/v1/payment-transactions/trx123/payer", "GET /acquisition/api/public/v1/payment-transactions/trx123"}; strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestPayTechnicalErrorRecoversAuthorizedTransaction(t *testing.T) {
	statusCalls := 0
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Simulate a technical error from the ANCV server: the client should
			// recover by checking the transaction status.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		statusCalls++
		_, _ = io.WriteString(w, `{"transaction":{"id":"trx123","state":"AUTHORIZED"}}`)
	})

	resp, err := client.pay("trx123", PayRequest{
		Payer: payer{BeneficiaryID: "1536923388807", Amount: Amount{Total: 3500, Currency: CurrencyEUR}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if statusCalls != 1 || resp.Transaction.State != StateAuthorized {
		t.Fatalf("status check should recover the payment: calls=%d, resp=%+v", statusCalls, resp)
	}
}

func TestAPIError(t *testing.T) {
	client := newClientWithMockANCVServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"errorCode":"INSUFFICIENT_BALANCE","errorMessage":"The balance is insufficient"}`)
	})

	_, err := client.GetTransaction("trx123")
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Code != "INSUFFICIENT_BALANCE" {
		t.Fatalf("unexpected API error: %+v", apiErr)
	}
	if !strings.Contains(err.Error(), "solde du compte") {
		t.Fatalf("error should be translated: %s", err)
	}
}
