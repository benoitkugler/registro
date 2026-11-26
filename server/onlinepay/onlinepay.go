// Package onlineay implements a simple online payment
// session, with support for differents providers.
// The common steps are :
//   - Init a payment session (or gateway), returning an opaque URL
//     to be access by the payer
//   - Listen for webhook notifications and store the
//     paiement in our DB
package onlinepay

import (
	"fmt"
	"io"
	"net/http"

	"registro/config"
	ds "registro/sql/dossiers"
)

// GatewayStart stores the input parameters required
// to start a payment session
type GatewayStart struct {
	IdDossier ds.IdDossier

	// [ResponsableMail] is either the responsable mail, or
	// it is empty and must be filled in the provider's page.
	ResponsableMail string

	Montant ds.Montant

	// Links to our espaceperso pages
	SuccesURL, CancelURL string
}

// Provider is either Stripe or Payrexx
type Provider interface {
	// InitGateway returns a unique gateway URL,
	// where we shoud redirect our user.
	InitGateway(params GatewayStart) (URL string, _ error)

	// ReceivePayment decodes the request send by the provider when a payment is concluded.
	// It returns [false] for other notifications, or an error if the
	// request is invalid.
	// [db] is only read, the returned [ds.Paiement] is not written yet.
	ReceivePayment(db ds.DB, body io.ReadCloser, headers http.Header) (ds.Paiement, bool, error)
}

var (
	_ Provider = stripeApi{}
	_ Provider = payrexx{}
)

// NewProvider builds the provider selected by the association config.
// It loads only the credentials needed by that provider.
func NewProvider(name config.OnlinePayProvider) (Provider, error) {
	switch name {
	case config.OnlinePayStripe:
		cfg, err := config.NewStripe()
		if err != nil {
			return nil, err
		}
		return stripeApi(cfg), nil
	case config.OnlinePayPayrexx:
		cfg, err := config.NewPayrexx()
		if err != nil {
			return nil, err
		}
		return payrexx(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported online payment provider %q", name)
	}
}
