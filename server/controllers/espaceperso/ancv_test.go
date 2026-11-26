package espaceperso

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/onlinepay/ancv"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"

	"github.com/labstack/echo/v4"
)

type ancvInitArgs struct {
	orderID         string
	montantAsCent   int
	notificationURL string
	returnContext   string
}

type stubANCVClient struct {
	statusResponse ancv.Response

	lastInitArgs ancvInitArgs
	lastPay      ancv.PayRequest
	statusCalls  int
}

func (s *stubANCVClient) NewInitRequest(orderID string, montantAsCent int, notificationURL, returnContext string) (ancv.InitRequest, error) {
	s.lastInitArgs = ancvInitArgs{
		orderID:         orderID,
		montantAsCent:   montantAsCent,
		notificationURL: notificationURL,
		returnContext:   returnContext,
	}
	return ancv.InitRequest{Order: ancv.Order{
		ID:        orderID,
		PaymentID: "payment-123",
		Amount:    ancv.Amount{Total: montantAsCent, Currency: ancv.CurrencyEUR},
	}}, nil
}

func (s *stubANCVClient) CreateAndPay(req ancv.InitRequest, beneficiaryID string) (ancv.PaymentResult, error) {
	s.lastPay = ancv.NewPayRequest(beneficiaryID, req.Order.Amount.Total)
	return ancv.PaymentResult{TransactionID: "trx123", State: "PROCESSING"}, nil
}

func (s *stubANCVClient) GetTransaction(id string) (ancv.Response, error) {
	s.statusCalls++
	if s.statusResponse.Transaction == nil {
		return ancv.Response{Transaction: &ancv.Transaction{ID: id, State: "REJECTED"}}, nil
	}
	return s.statusResponse, nil
}

func setupANCVDossier(t *testing.T) (tu.TestDB, *Controller, string, ds.IdDossier) {
	t.Helper()
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	t.Cleanup(db.Remove)

	person, err := pr.Personne{Identite: pr.Identite{Prenom: "Jean", Nom: "DUPONT"}}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: person.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := cps.Camp{IdTaux: 1, Prix: ds.NewEuros(400)}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{
		IdTaux: 1, IdCamp: camp.Id, IdPersonne: person.Id, IdDossier: dossier.Id, Statut: cps.Inscrit,
	}.Insert(db)
	tu.AssertNoErr(t, err)

	key := crypto.NewEncrypter("test-key")
	ct := Controller{
		db:  db.DB,
		key: key,
		asso: config.Asso{
			SupportANCV:    true,
			AcomptePercent: 0,
		},
	}
	return db, &ct, crypto.EncryptID(key, dossier.Id), dossier.Id
}

func ancvContext(body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ancv/notification", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestStartANCVPayment(t *testing.T) {
	_, ct, token, idDossier := setupANCVDossier(t)
	stub := &stubANCVClient{}
	ct.ancvClient = stub

	body, err := json.Marshal(StartANCVPaymentIn{
		Token: token, BeneficiaryID: "1536923388807", MontantAsCent: 3500,
	})
	tu.AssertNoErr(t, err)
	c, rec := ancvContext(string(body))
	tu.AssertNoErr(t, ct.StartANCVPayment(c))

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d (%s)", rec.Code, rec.Body.String())
	}
	var out StartANCVPaymentOut
	tu.AssertNoErr(t, json.Unmarshal(rec.Body.Bytes(), &out))
	if out.TransactionID != "trx123" || out.State != "PROCESSING" {
		t.Fatalf("unexpected output: %+v", out)
	}

	init := stub.lastInitArgs
	if init.orderID != strconv.FormatInt(int64(idDossier), 10) {
		t.Fatalf("unexpected order id: %s", init.orderID)
	}
	if init.montantAsCent != 3500 {
		t.Fatalf("unexpected amount: %d", init.montantAsCent)
	}
	if init.returnContext != token {
		t.Fatal("returnContext should carry the dossier token")
	}
	if !strings.HasSuffix(init.notificationURL, "/api/v1/ancv/notification") {
		t.Fatalf("unexpected webhook URL: %s", init.notificationURL)
	}
	if want := ancv.NewPayRequest("1536923388807", 3500); stub.lastPay != want {
		t.Fatalf("unexpected payer call: got %+v, want %+v", stub.lastPay, want)
	}
}

func TestStartANCVPaymentValidation(t *testing.T) {
	_, ct, token, _ := setupANCVDossier(t)
	ct.ancvClient = (&stubANCVClient{})

	cases := []StartANCVPaymentIn{
		{Token: token, BeneficiaryID: "1536923388807", MontantAsCent: 0},
		{Token: token, BeneficiaryID: "1536923388807", MontantAsCent: 40001},
		{Token: "invalid", BeneficiaryID: "1536923388807", MontantAsCent: 3500},
	}
	for _, args := range cases {
		body, err := json.Marshal(args)
		tu.AssertNoErr(t, err)
		c, _ := ancvContext(string(body))
		if err := ct.StartANCVPayment(c); err == nil {
			t.Fatalf("expected an error for %+v", args)
		}
	}
}

func authorizedNotification(token string, id ds.IdDossier, state string) ancv.Response {
	return ancv.Response{
		Transaction: &ancv.Transaction{
			ID:    "trx123",
			State: state,
			Order: ancv.Order{
				ID:        strconv.FormatInt(int64(id), 10),
				PaymentID: "payment-123",
				Amount:    ancv.Amount{Total: 4000, Currency: ancv.CurrencyEUR},
			},
			Payers: []ancv.PayerInfo{{
				Authorizations: []ancv.Authorization{{
					Type:   "CVCo",
					Amount: ancv.Amount{Total: 3500, Currency: ancv.CurrencyEUR},
				}},
			}},
		},
		ApplicationContext: ancv.ApplicationContext{ReturnContext: token},
	}
}

func TestHandleANCVPaymentNotification(t *testing.T) {
	db, ct, token, idDossier := setupANCVDossier(t)
	stub := &stubANCVClient{statusResponse: authorizedNotification(token, idDossier, ancv.StateAuthorized)}
	ct.ancvClient = stub

	c, rec := ancvContext(`{"transaction":{"id":"trx123"},"responseDate":"2018-08-28T10:18:00Z"}`)
	tu.AssertNoErr(t, ct.HandleANCVPaymentNotification(c))
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d (%s)", rec.Code, rec.Body.String())
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(db.DB, idDossier)
	tu.AssertNoErr(t, err)
	if len(paiements) != 1 {
		t.Fatalf("expected one payment, got %d", len(paiements))
	}
	var p ds.Paiement
	for _, paiement := range paiements {
		p = paiement
	}
	if p.Mode != ds.Ancv || p.Label != "trx123" || p.Details != "payment-123" ||
		p.Montant != (ds.Montant{Cent: 3500, Currency: ds.Euros}) {
		t.Fatalf("unexpected payment: %+v", p)
	}

	// ANCV retries up to 4 times (Kit §5.1.2.1, p. 45) : a second identical
	// notification must be ignored.
	c, rec = ancvContext(`{"transaction":{"id":"trx123"}}`)
	tu.AssertNoErr(t, ct.HandleANCVPaymentNotification(c))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("duplicate should be ignored, got %d", rec.Code)
	}
	paiements, err = ds.SelectPaiementsByIdDossiers(db.DB, idDossier)
	tu.AssertNoErr(t, err)
	if len(paiements) != 1 {
		t.Fatalf("duplicate payment was inserted: %d", len(paiements))
	}
}

func TestHandleANCVPaymentNotificationIgnoredStates(t *testing.T) {
	db, ct, token, idDossier := setupANCVDossier(t)
	stub := &stubANCVClient{statusResponse: authorizedNotification(token, idDossier, "REJECTED")}
	ct.ancvClient = stub

	c, rec := ancvContext(`{"transaction":{"id":"trx123"}}`)
	tu.AssertNoErr(t, ct.HandleANCVPaymentNotification(c))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("rejected transaction should be ignored, got %d", rec.Code)
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(db.DB, idDossier)
	tu.AssertNoErr(t, err)
	if len(paiements) != 0 {
		t.Fatalf("no payment should be inserted, got %d", len(paiements))
	}

	// A pre-transaction notification is not part of this flow.
	c, rec = ancvContext(`{"preTransaction":{"id":"pre123"}}`)
	tu.AssertNoErr(t, ct.HandleANCVPaymentNotification(c))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("pre-transaction should be ignored, got %d", rec.Code)
	}
}

func ancvStatusContext(token, transactionID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/espaceperso/paiement/ancv/status", nil)
	query := req.URL.Query()
	query.Set("token", token)
	query.Set("transactionId", transactionID)
	req.URL.RawQuery = query.Encode()
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestGetANCVPaymentStatus(t *testing.T) {
	db, ct, token, idDossier := setupANCVDossier(t)
	ct.ancvClient = &stubANCVClient{}

	status := func(token, transactionID string) ANCVPaymentStatusOut {
		t.Helper()
		c, rec := ancvStatusContext(token, transactionID)
		tu.AssertNoErr(t, ct.GetANCVPaymentStatus(c))
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d (%s)", rec.Code, rec.Body.String())
		}
		var out ANCVPaymentStatusOut
		tu.AssertNoErr(t, json.Unmarshal(rec.Body.Bytes(), &out))
		return out
	}

	if status(token, "trx123").Recorded {
		t.Fatal("payment should not be recorded before the notification")
	}

	_, err := ds.Paiement{
		IdDossier: idDossier,
		Payeur:    "Jean DUPONT",
		Mode:      ds.Ancv,
		Montant:   ds.Montant{Cent: 3500, Currency: ds.Euros},
		Time:      time.Now().Truncate(time.Second),
		Label:     "trx123",
		Details:   "payment-123",
	}.Insert(db.DB)
	tu.AssertNoErr(t, err)

	if !status(token, "trx123").Recorded {
		t.Fatal("recorded payment should be detected")
	}
	if status(token, "other").Recorded {
		t.Fatal("another transaction should not be reported as recorded")
	}
	if status(crypto.EncryptID(ct.key, idDossier+1), "trx123").Recorded {
		t.Fatal("payment from another dossier should not be reported")
	}
}
