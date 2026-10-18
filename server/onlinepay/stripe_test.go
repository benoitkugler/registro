package onlinepay

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"

	"registro/config"
	"registro/sql/dossiers"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestStripeStart(t *testing.T) {
	tu.LoadEnv(t, "../env.sh")
	cfg, err := config.NewStripe()
	tu.AssertNoErr(t, err)

	api := stripeApi(cfg)
	id, err := api.StartGateway(GatewayStart{
		IdDossier: 1,
		Montant:   dossiers.NewFrancsuisses(2.53),
		SuccesURL: "http://test.fr",
		CancelURL: "http://test.fr",
	})
	tu.AssertNoErr(t, err)
	fmt.Println(id)
}

type stripePayload struct {
	Payload   []byte
	Signature string
}

// func TestStripeReceivePaiement(t *testing.T) {
// 	var pEur, pChf stripePayload
// 	fEur, err := os.ReadFile("stripedev/test_eur.json")
// 	tu.AssertNoErr(t, err)
// 	err = json.Unmarshal(fEur, &pEur)
// 	tu.AssertNoErr(t, err)
// 	fChf, err := os.ReadFile("stripedev/test_chf.json")
// 	tu.AssertNoErr(t, err)
// 	err = json.Unmarshal(fChf, &pChf)
// 	tu.AssertNoErr(t, err)

// 	tu.LoadEnv(t, "../env.sh")
// 	cfg, err := config.NewStripe()
// 	tu.AssertNoErr(t, err)

// 	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
// 		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
// 		"../migrations/init.sql")
// 	defer db.Remove()

// 	p, err := pr.Personne{}.Insert(db)
// 	tu.AssertNoErr(t, err)
// 	d, err := ds.Dossier{IdTaux: 1, IdResponsable: p.Id}.Insert(db)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, d.Id == 1)

// 	api := stripeApi(cfg)

// 	paiementEur, ok, err := api.receivePaiement(db, []byte(pEur.Payload), pEur.Signature)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, ok)
// 	tu.Assert(t, paiementEur.Montant == ds.Montant{Cent: 3000, Currency: ds.Euros})
// 	tu.Assert(t, !paiementEur.IsRemboursement)

// 	paiementChf, ok, err := api.receivePaiement(db, []byte(pChf.Payload), pChf.Signature)
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, ok)
// 	tu.Assert(t, paiementChf.Montant == ds.Montant{Cent: 3000, Currency: ds.FrancsSuisse})
// 	tu.Assert(t, !paiementChf.IsRemboursement)

// 	_, _, err = api.receivePaiement(db, nil, "")
// 	tu.AssertErr(t, err)
// 	_, _, err = api.receivePaiement(db, []byte(pChf.Payload), "") // missing signature
// 	tu.AssertErr(t, err)
// }

func TestStripeReceivePaiement(t *testing.T) {
	// this test requires stripe CLI forwarding, setup with
	// stripe listen --all-thin --forward-to localhost:1324/stripe_test -l

	// initial setup
	tu.LoadEnv(t, "../env.sh")
	cfg, err := config.NewStripe()
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	p, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	d, err := ds.Dossier{IdTaux: 1, IdResponsable: p.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	tu.Assert(t, d.Id == 1)

	api := stripeApi(cfg)

	events := make(chan stripePayload)
	http.HandleFunc("/stripe_test", func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("Stripe-Signature")
		b, _ := io.ReadAll(r.Body)
		events <- stripePayload{b, signature}
	})

	go func() {
		fmt.Println("Setting up go test endpoint...")
		err := http.ListenAndServe(":1324", nil)
		if err != nil {
			panic(err)
		}
	}()

	testCurrency := func(t *testing.T, c ds.Currency) {
		go func() {
			currency := c.IsoLower()
			fmt.Println("Triggering event with currency", currency, "...")
			err := exec.Command("stripe", "trigger", "checkout.session.completed", "--override", fmt.Sprintf("price:currency=%s", currency), "--override", fmt.Sprintf("checkout_session:currency=%s", currency), "--override", "checkout_session:metadata.id_dossier=1").Run()
			if err != nil {
				panic(err)
			}
		}()

		for {
			event, hasEvent := <-events
			fmt.Println("hasEvent", hasEvent)
			isSessionCompleted := bytes.Contains(event.Payload, []byte("checkout.session.completed"))
			paiement, ok, err := api.receivePaiement(db, event.Payload, event.Signature)
			tu.AssertNoErr(t, err) // all event should success
			tu.Assert(t, ok == isSessionCompleted)
			fmt.Println("checking event")
			if isSessionCompleted {
				tu.Assert(t, paiement.Montant == ds.Montant{Cent: 3000, Currency: c})
				tu.Assert(t, !paiement.IsRemboursement)
			}
		}
	}
	testCurrency(t, ds.Euros)
	// testCurrency(t, ds.FrancsSuisse, 1325)
}
