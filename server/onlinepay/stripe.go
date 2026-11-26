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

// InitGateway implements [Provider]. It starts a Stripe Checkout session.
func (config stripeApi) InitGateway(start GatewayStart) (string, error) {
	stripe.Key = config.Key

	// only fill if not using custom mail
	var customerEmail *string
	if start.ResponsableMail != "" {
		customerEmail = stripe.String(start.ResponsableMail)
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
					Currency: stripe.String(start.Montant.Currency.IsoLower()),
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
	return out.URL, nil
}

// ReceivePayment implements [Provider]. It decodes the request send by Stripe
// when a payment is concluded, and returns [false] for other notifications.
func (config stripeApi) ReceivePayment(db ds.DB, body io.ReadCloser, headers http.Header) (ds.Paiement, bool, error) {
	payload, err := io.ReadAll(body)
	if err != nil {
		return ds.Paiement{}, false, fmt.Errorf("error reading request body from Stripe notification: %v", err)
	}
	_ = body.Close()

	stripeSignature := headers.Get("Stripe-Signature")

	stripe.Key = config.Key

	client := stripe.NewClient(config.Key)
	// Pass the request body and Stripe-Signature header to ParseEventNotification, along with the webhook signing key
	// You can find your endpoint's secret in your webhook settings
	event, err := client.ParseEventNotification(payload, stripeSignature, config.Webhook)
	if err != nil {
		return ds.Paiement{}, false, fmt.Errorf("invalid request on Stripe notification endpoint: %v", err)
	}

	// Handle the checkout.session.completed notif
	notif, ok := event.(*stripe.V1CheckoutSessionCompletedEventNotification)
	if !ok {
		return ds.Paiement{}, false, nil
	}

	session, err := notif.FetchRelatedObject(context.Background())
	if err != nil {
		return ds.Paiement{}, false, fmt.Errorf("failed to retrieve Stripe event: %s", err)
	}
	paiement, err := parseStripePaiement(db, session)
	if err != nil {
		return ds.Paiement{}, false, err
	}

	// Providers may retry the same notification: avoid creating the
	// paiement twice.
	existing, err := ds.SelectPaiementsByIdDossiers(db, paiement.IdDossier)
	if err != nil {
		return ds.Paiement{}, false, utils.SQLError(err)
	}
	for _, p := range existing {
		if p.Label == paiement.Label && p.Mode == paiement.Mode && p.Montant == paiement.Montant {
			// ignore without error
			return ds.Paiement{}, false, nil
		}
	}

	return paiement, true, err
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
