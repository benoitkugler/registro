package onlinepay

import (
	"errors"
	"io"
	"net/http"

	"registro/config"
	ds "registro/sql/dossiers"
)

// https://docs.payrexx.com/developer/guides/gateway
type payrexx config.Payrexx

// InitGateway implements [Provider]. Payrexx is selected by the Repère
// configuration, but the gateway API is not implemented yet.
func (api payrexx) InitGateway(params GatewayStart) (id string, _ error) {
	return "", errors.New("Payrexx online payment is not implemented")
}

// ReceivePayment implements [Provider].
func (api payrexx) ReceivePayment(db ds.DB, body io.ReadCloser, headers http.Header) (ds.Paiement, bool, error) {
	return ds.Paiement{}, false, errors.New("Payrexx online payment is not implemented")
}
