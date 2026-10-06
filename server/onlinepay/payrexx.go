package onlinepay

import "registro/config"

// https://docs.payrexx.com/developer/guides/gateway
type payrexx config.Payrexx

// StartSession should be called to start a paiement session.
func (api payrexx) InitGateway(params GatewayStart) (id string, _ error) {
	// TODO:
	return "", nil
}
