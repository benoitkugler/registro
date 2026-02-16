package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"registro/config"
	"registro/sql/dossiers"
	"registro/sql/personnes"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
)

// entrées des metadatas
const metadataJSON = "metadata_json"

type stripeMetadata struct {
	IdDossier dossiers.IdDossier
	Payeur    string
	Montant   dossiers.Montant
}

// StartSession should be called to start a paiement session.
func StartSession(key config.Stripe, idDossier dossiers.IdDossier, respo personnes.Identite, userProvidedMail bool,
	montant dossiers.Montant,
	succesURL, cancelURL string,
) (sessionID string, _ error) {
	stripe.Key = key.Key

	// only fill if not using custom mail
	var customerEmail *string
	if !userProvidedMail {
		customerEmail = stripe.String(respo.Mail)
	}

	md := stripeMetadata{
		IdDossier: idDossier,
		Payeur:    respo.NOMPrenom(),
		Montant:   montant,
	}
	mdJSON, err := json.Marshal(md)
	if err != nil {
		return "", err
	}

	var currency string
	switch montant.Currency {
	case dossiers.Euros:
		currency = "eur"
	case dossiers.FrancsSuisse:
		currency = "chf"
	default:
		return "", fmt.Errorf("unsupported currency: %d", montant.Currency)
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			metadataJSON: string(mdJSON),
		},
		CustomerEmail: customerEmail,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Séjour de vacances"),
					},
					UnitAmount: stripe.Int64(int64(montant.Cent)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: &succesURL,
		CancelURL:  &cancelURL,
	}
	session, err := session.New(params)
	if err != nil {
		return "", err
	}

	return session.ID, nil
}

// ReceivePaiement decodes the request send by Stripe when a payment is concluded.
// It returns [false] for other notifications.
func ReceivePaiement(key config.Stripe, body io.ReadCloser, header http.Header) (dossiers.Paiement, bool, error) {
	stripe.Key = key.Key

	payload, err := io.ReadAll(body)
	if err != nil {
		return dossiers.Paiement{}, false, fmt.Errorf("error reading request body from Stripe notification: %v", err)
	}
	_ = body.Close()

	// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key
	// You can find your endpoint's secret in your webhook settings
	event, err := webhook.ConstructEvent(payload, header.Get("Stripe-Signature"), key.Webhook)
	if err != nil {
		return dossiers.Paiement{}, false, fmt.Errorf("invalid request on Stripe notification endpoint: %v", err)
	}

	// Handle the checkout.session.completed event
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return dossiers.Paiement{}, false, fmt.Errorf("error parsing webhook JSON: %v", err)
		}
		p, err := parsePaiement(&session)
		return p, true, err
	}

	return dossiers.Paiement{}, false, nil
}

// parsePaiement builds a [dossiers.Paiement] object from a Stripe notification
func parsePaiement(session *stripe.CheckoutSession) (dossiers.Paiement, error) {
	var md stripeMetadata
	err := json.Unmarshal([]byte(session.Metadata[metadataJSON]), &md)
	if err != nil {
		return dossiers.Paiement{}, err
	}

	paiement := dossiers.Paiement{
		IdDossier: md.IdDossier,
		Payeur:    md.Payeur,
		Mode:      dossiers.EnLigne,
		Label:     session.PaymentIntent.ID,
		Montant:   md.Montant,
		Time:      time.Now().Truncate(time.Second),
	}

	return paiement, nil
}
