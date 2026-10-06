// Package onlineay implements a simple online payment
// session, with support for differents providers.
// The common steps are :
//   - Init a payment session, returning an opaque URL
//     to be access by the payer
//   - Listen for webhook notifications and store the
//     paiement in our DB
package onlinepay

import (
	"io"
	"net/http"

	"registro/sql/dossiers"
	"registro/sql/personnes"
)

// GatewayStart stores the input parameters required
// to start a payment session
type GatewayStart struct {
	IdDossier   dossiers.IdDossier
	Responsable personnes.Identite
	// If true, the [Responsable.Mail] is used,
	// otherwise it is empty and must be filled in the provider page.
	UserProvidedMail bool
	Montant          dossiers.Montant
	// Links to our espaceperso pages
	SuccesURL, CancelURL string
}

// Provider is either Stripe or Payrexx
type Provider interface {
	// InitGateway returns a unique gateway ID,
	// which should be sufficient for the frontend
	// to construct the payement URL.
	InitGateway(params GatewayStart) (id string, _ error)

	// ReceivePayment decodes the request send by the provider when a payment is concluded.
	// It returns [false] for other notifications, or an error if the
	// request is invalid.
	ReceivePayment(body io.ReadCloser, headers http.Header) (dossiers.Paiement, bool, error)
}
