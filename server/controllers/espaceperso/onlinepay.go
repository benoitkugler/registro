package espaceperso

import (
	"errors"
	"fmt"
	"net/http"

	"registro/crypto"
	"registro/logic"
	"registro/onlinepay"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// minOnlinePaymentCent is the minimum amount accepted for an online payment.
// It is expressed in each currency's minor unit (50 €, 50 CHF).
const minOnlinePaymentCent = 5000

// PaiementOption describes the amount still to pay in one currency,
// together with the smallest payment accepted in that currency.
type PaiementOption struct {
	Currency ds.Currency
	Min, Max int // cents
}

func (po PaiementOption) montants() (min, max ds.Montant) {
	return ds.Montant{Currency: po.Currency, Cent: po.Min}, ds.Montant{Currency: po.Currency, Cent: po.Max}
}

// paiementOptions converts [montant] into every currency supported by the
// dossier. Online payments may be split, so the minimum is clamped to the
// maximum when the amount asked is smaller than 50 units.
func paiementOptions(taux ds.Taux, montant ds.Montant) map[ds.Currency]PaiementOption {
	convertible := taux.Convertible(montant)
	currencies := taux.SupportedCurrencies()
	out := make(map[ds.Currency]PaiementOption, len(currencies))
	for _, currency := range currencies {
		max := convertible.Convert(currency)
		out[currency] = PaiementOption{
			Currency: currency,
			Max:      max.Cent,
			Min:      min(minOnlinePaymentCent, max.Cent),
		}
	}
	return out
}

// StartOnlinePaymentIn is sent by the family. The amount is always checked
// against the dossier: the client cannot choose an arbitrary price.
type StartOnlinePaymentIn struct {
	Token   string
	Montant ds.Montant
}

// StartOnlinePaymentOut contains the gateway URL to redirect to.
type StartOnlinePaymentOut struct {
	URL string
}

// StartOnlinePayment starts a payment session with the configured provider.
func (ct *Controller) StartOnlinePayment(c echo.Context) error {
	if !ct.asso.SupportOnlinePay || ct.provider == nil {
		return errors.New("Le paiement en ligne n'est pas disponible.")
	}

	var args StartOnlinePaymentIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return err
	}

	dossier, err := logic.LoadDossiersFinance(ct.db, idDossier)
	if err != nil {
		return err
	}
	montantARegler, ok := dossier.MontantARegler(ct.asso.AcomptePercent, ct.asso.RequireDemandePaiement)
	if !ok {
		return errors.New("Aucun paiement n'est demandé pour ce dossier.")
	}

	option, found := paiementOptions(dossier.Taux, montantARegler)[args.Montant.Currency]
	if !found {
		return fmt.Errorf("devise non supportée pour ce dossier : %d", args.Montant.Currency)
	}
	if args.Montant.Cent < option.Min || args.Montant.Cent > option.Max {
		min, max := option.montants()
		return fmt.Errorf("le montant doit être compris entre %s et %s", min, max)
	}

	start := onlinepay.GatewayStart{
		IdDossier:       idDossier,
		ResponsableMail: dossier.Responsable().Mail,
		Montant:         args.Montant,
		SuccesURL: logic.EspacePersoURL(
			ct.key, c.Request().Host, idDossier, utils.QP("paiement", "success"),
		),
		CancelURL: logic.EspacePersoURL(
			ct.key, c.Request().Host, idDossier, utils.QP("paiement", "cancel"),
		),
	}
	url, err := ct.provider.InitGateway(start)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, StartOnlinePaymentOut{URL: url})
}

// HandleOnlinePaymentNotification is the public webhook called by the
// payment provider. Completed payments are stored as ds.Paiement entries.
func (ct *Controller) HandleOnlinePaymentNotification(c echo.Context) error {
	if ct.provider == nil {
		return errors.New("Le paiement en ligne n'est pas disponible.")
	}

	paiement, ok, err := ct.provider.ReceivePayment(ct.db, c.Request().Body, c.Request().Header)
	if err != nil {
		return err
	}
	if !ok {
		return c.NoContent(http.StatusNoContent)
	}

	if _, err := paiement.Insert(ct.db); err != nil {
		return utils.SQLError(err)
	}

	return c.NoContent(http.StatusOK)
}
