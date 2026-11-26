package espaceperso

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"registro/crypto"
	"registro/logic"
	"registro/onlinepay/ancv"
	ds "registro/sql/dossiers"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// ancvClient is the subset of the ANCV API used by the controller. It is an
// interface so that handlers can be tested with a stub.
type ancvClient interface {
	NewInitRequest(orderID string, montantAsCent int, notificationURL, returnContext string) (ancv.InitRequest, error)
	CreateAndPay(ancv.InitRequest, string) (ancv.PaymentResult, error)
	GetTransaction(string) (ancv.Response, error)
}

// StartANCVPaymentIn is sent by the family. ANCV only supports euros, so the
// amount is given in cents without a currency field.
type StartANCVPaymentIn struct {
	Token string // encrypted dossier token, also used as returnContext

	BeneficiaryID string // ANCV account number, or the email address linked to it
	MontantAsCent int    // euros, in cents
}

// StartANCVPaymentOut describes the transaction submitted to ANCV.
type StartANCVPaymentOut struct {
	TransactionID string
	State         string // PROCESSING while the beneficiary validates
}

// StartANCVPayment starts a payment through ANCV.
//
// Kit d'intégration ANCV v1.08, §4.5.1 (p. 17) : « La première étape
// du paiement [...] est l'initialisation d'une transaction de paiement ».
// §4.5.2 (p. 20) : « La mise en paiement consiste à soumettre une transaction
// [...] pour acquisition du paiement par un bénéficiaire ». Le paiement est
// ensuite validé par le bénéficiaire sur son application ; la transaction est
// enregistrée dans [HandleANCVPaymentNotification].
func (ct *Controller) StartANCVPayment(c echo.Context) error {
	if !ct.asso.SupportANCV || ct.ancvClient == nil {
		return errors.New("Le paiement par Chèque-Vacances Connect n'est pas disponible.")
	}

	var args StartANCVPaymentIn
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

	option, found := paiementOptions(dossier.Taux, montantARegler)[ds.Euros]
	if !found {
		return errors.New("Le paiement par Chèque-Vacances Connect n'est disponible qu'en euros.")
	}
	// Kit §4.5.1.3 (p. 17) : « Le montant minimum de transaction accepté est : 1ct ».
	if args.MontantAsCent < 1 || args.MontantAsCent > option.Max {
		return fmt.Errorf("le montant doit être compris entre 1 centime et %s", ds.Montant{Cent: option.Max, Currency: ds.Euros})
	}
	if len(args.Token) > 512 {
		return errors.New("contexte de paiement trop long")
	}

	redirectURL := utils.BuildUrl(c.Request().Host, "api/v1/ancv/notification")
	initReq, err := ct.ancvClient.NewInitRequest(
		strconv.FormatInt(int64(idDossier), 10),
		args.MontantAsCent,
		redirectURL,
		args.Token,
	)
	if err != nil {
		return err
	}

	payment, err := ct.ancvClient.CreateAndPay(initReq, args.BeneficiaryID)
	if err != nil {
		return err
	}

	out := StartANCVPaymentOut{TransactionID: payment.TransactionID, State: payment.State}
	return c.JSON(http.StatusOK, out)
}

// ANCVPaymentStatusOut reports whether an ANCV payment has been recorded for
// the dossier. It deliberately says nothing about the ANCV transaction state:
// only the webhook inserts the payment, so Recorded is the authoritative flag
// for the family's UI.
type ANCVPaymentStatusOut struct {
	Recorded bool
}

// GetANCVPaymentStatus answers the family's UI while it waits for the ANCV
// notification. Kit §5.1.2.1 (p. 45) : ANCV may retry the webhook, so the UI
// polls this lightweight endpoint instead of assuming success from the
// PROCESSING state returned by StartANCVPayment.
func (ct *Controller) GetANCVPaymentStatus(c echo.Context) error {
	if !ct.asso.SupportANCV || ct.ancvClient == nil {
		return errors.New("Le paiement par Chèque-Vacances Connect n'est pas disponible.")
	}
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, c.QueryParam("token"))
	if err != nil {
		return errors.New("Lien invalide.")
	}
	transactionID := c.QueryParam("transactionId")
	if transactionID == "" {
		return errors.New("Identifiant de transaction ANCV manquant.")
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(ct.db, idDossier)
	if err != nil {
		return utils.SQLError(err)
	}
	for _, paiement := range paiements {
		if paiement.Mode == ds.Ancv && !paiement.IsRemboursement && paiement.Label == transactionID {
			return c.JSON(http.StatusOK, ANCVPaymentStatusOut{Recorded: true})
		}
	}
	return c.JSON(http.StatusOK, ANCVPaymentStatusOut{})
}

// HandleANCVPaymentNotification is the public webhook called by ANCV.
//
// Kit §5.1.1 (p. 45) : « Ces services doivent être en capacité de traiter une
// requête avec transmission de paramètres par méthode POST en REST/JSON ».
// Kit §5.1.2.1 (p. 45) prévoit jusqu'à 4 tentatives supplémentaires en cas
// d'échec : le traitement est donc idempotent.
//
// Le corps du webhook n'est pas signé (§5.1.4, p. 46). L'http.Header et le
// montant sont donc relus depuis l'API de demande d'état (§4.5.3, p. 23) avant
// d'enregistrer quoi que ce soit.
func (ct *Controller) HandleANCVPaymentNotification(c echo.Context) error {
	if ct.ancvClient == nil {
		return errors.New("Le paiement par Chèque-Vacances Connect n'est pas disponible.")
	}

	var notification ancv.Response
	if err := c.Bind(&notification); err != nil {
		return err
	}
	// Les notifications de pré-transaction ne concernent pas ce parcours.
	// Kit §5.1.3 (p. 46) : « transaction [...] seulement présent si l'objet
	// pre-transaction n'est pas transmis ».
	if notification.Transaction == nil || len(notification.PreTransaction) != 0 {
		return c.NoContent(http.StatusNoContent)
	}

	verified, err := ct.ancvClient.GetTransaction(notification.Transaction.ID)
	if err != nil {
		return err // 400 : ANCV re-tentera l'appel
	}
	transaction := verified.Transaction
	if transaction == nil || !ancv.IsAuthorizedState(transaction.State) {
		return c.NoContent(http.StatusNoContent)
	}

	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, verified.ApplicationContext.ReturnContext)
	if err != nil {
		return fmt.Errorf("contexte de paiement ANCV invalide : %w", err)
	}
	if transaction.Order.ID != strconv.FormatInt(int64(idDossier), 10) {
		return errors.New("incohérence entre la transaction ANCV et le dossier")
	}
	if transaction.Order.PaymentID == "" || transaction.Order.Amount.Currency != ancv.CurrencyEUR {
		return errors.New("transaction ANCV invalide")
	}

	// Kit §6.2.2.19 (pp. 62-63) : l'autorisation peut être d'un montant
	// inférieur à celui demandé ; c'est le montant réellement autorisé qui est
	// encaissé. On n'enregistre que les autorisations ANCV.
	amount := 0
	for _, payer := range transaction.Payers {
		for _, authorization := range payer.Authorizations {
			if authorization.Type == "CVCo" && authorization.Amount.Currency == ancv.CurrencyEUR {
				amount += authorization.Amount.Total
			}
		}
	}
	if amount <= 0 || amount > transaction.Order.Amount.Total {
		return fmt.Errorf("montant autorisé ANCV invalide : %d", amount)
	}

	existing, err := ds.SelectPaiementsByIdDossiers(ct.db, idDossier)
	if err != nil {
		return utils.SQLError(err)
	}
	for _, paiement := range existing {
		if paiement.Label == transaction.ID && paiement.Mode == ds.Ancv &&
			paiement.Montant == (ds.Montant{Cent: amount, Currency: ds.Euros}) {
			return c.NoContent(http.StatusNoContent) // notification déjà traitée
		}
	}

	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	paiement := ds.Paiement{
		IdDossier: idDossier,
		Payeur:    dossier.Responsable().NOMPrenom(),
		Mode:      ds.Ancv,
		Montant:   ds.Montant{Cent: amount, Currency: ds.Euros},
		Time:      time.Now().Truncate(time.Second),
		Label:     transaction.ID,
		Details:   transaction.Order.PaymentID,
	}
	if _, err := paiement.Insert(ct.db); err != nil {
		return utils.SQLError(err)
	}
	return c.NoContent(http.StatusOK)
}
