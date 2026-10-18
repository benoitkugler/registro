package onlinepay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"registro/config"
	"registro/logic"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/checkout/session"
)

type stripeApi config.Stripe

// entrées des metadatas
const metadataKey = "id_dossier"

// StartGateway should be called to start a paiement session.
func (config stripeApi) StartGateway(start GatewayStart) (sessionID string, _ error) {
	stripe.Key = config.Key

	// only fill if not using custom mail
	var customerEmail *string
	if start.ResponsableMail != "" {
		customerEmail = stripe.String(start.ResponsableMail)
	}

	var currency string
	switch start.Montant.Currency {
	case ds.Euros:
		currency = "eur"
	case ds.FrancsSuisse:
		currency = "chf"
	default:
		return "", fmt.Errorf("unsupported currency: %d", start.Montant.Currency)
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			metadataKey: strconv.Itoa(int(start.IdDossier)),
		},
		CustomerEmail: customerEmail,
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Séjour de vacances"),
					},
					UnitAmount: stripe.Int64(int64(start.Montant.Cent)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(start.SuccesURL),
		CancelURL:  stripe.String(start.CancelURL),
	}
	out, err := session.New(params)
	if err != nil {
		return "", err
	}
	return out.ID, nil
}

// ReceivePaiement implements [Provider]
func (config stripeApi) ReceivePaiement(db ds.DB, body io.ReadCloser, headers http.Header) (ds.Paiement, bool, error) {
	payload, err := io.ReadAll(body)
	if err != nil {
		return ds.Paiement{}, false, fmt.Errorf("error reading request body from Stripe notification: %v", err)
	}
	_ = body.Close()

	stripeSignature := headers.Get("Stripe-Signature")

	return config.receivePaiement(db, payload, stripeSignature)
}

// receivePaiement decodes the request send by Stripe when a payment is concluded.
// It returns [false] for other notifications.
func (config stripeApi) receivePaiement(db ds.DB, payload []byte, stripeSignature string) (ds.Paiement, bool, error) {
	stripe.Key = config.Key

	client := stripe.NewClient(config.Key)
	// Pass the request body and Stripe-Signature header to ParseEventNotification, along with the webhook signing key
	// You can find your endpoint's secret in your webhook settings
	event, err := client.ParseEventNotification(payload, stripeSignature, config.Webhook)
	if err != nil {
		return ds.Paiement{}, false, fmt.Errorf("invalid request on Stripe notification endpoint: %v", err)
	}

	// Handle the checkout.session.completed notif
	if notif, ok := event.(*stripe.V1CheckoutSessionCompletedEventNotification); ok {
		session, err := notif.FetchRelatedObject(context.Background())
		if err != nil {
			return ds.Paiement{}, false, fmt.Errorf("failed to retrieve Stripe event: %s", err)
		}
		p, err := parseStripePaiement(db, session)
		return p, true, err
	}

	return ds.Paiement{}, false, nil
}

// parseStripePaiement builds a [ds.Paiement] object from a Stripe notification
func parseStripePaiement(db ds.DB, session *stripe.CheckoutSession) (ds.Paiement, error) {
	currency, err := parseCurrency(string(session.Currency))
	if err != nil {
		return ds.Paiement{}, err
	}
	montant := int(session.AmountTotal)
	idDossier, err := utils.ParseInt[ds.IdDossier](session.Metadata[metadataKey])
	if err != nil {
		return ds.Paiement{}, err
	}
	dossier, err := logic.LoadDossier(db, idDossier)
	if err != nil {
		return ds.Paiement{}, err
	}

	paiement := ds.Paiement{
		IdDossier: dossier.Dossier.Id,
		Payeur:    dossier.Responsable().NOMPrenom(),
		Mode:      ds.EnLigne,
		Label:     session.PaymentIntent.ID,
		Montant:   ds.Montant{Cent: montant, Currency: currency},
		Time:      time.Now().Truncate(time.Second),
	}

	return paiement, nil
}

func parseCurrency(c string) (ds.Currency, error) {
	switch c {
	case "eur":
		return ds.Euros, nil
	case "chf":
		return ds.FrancsSuisse, nil
	default:
		return 0, fmt.Errorf("internal error: unsupported currency %s", c)
	}
}
