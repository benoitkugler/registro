// Implémente un relais Horizon <-> Helloasso
// exposant la fonction Checkout sans partager l'accès complet aux données HelloAsso
//
// TODO: vérifier urls retours
package horizon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"registro/config"
	"registro/helloasso"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// reference, type
const baseHorizonCallbackURL = "https://www.horizon-evenements.com/paiement/retour?ref=%s&type=%s"
const horizonNotificationURL = "https://www.horizon-evenements.com/api/helloasso/relais"

type Controller struct {
	Helloasso config.Helloasso
	Horizon   config.Horizon
}

func readBearer(header string) string {
	_, key, ok := strings.Cut(header, "Bearer")
	if !ok {
		return ""
	}
	return strings.TrimSpace(key)
}

func (ct *Controller) checkHorizonKey(c echo.Context) bool {
	key := readBearer(c.Request().Header.Get("Authorization"))
	return key == ct.Horizon.HorizonKey
}

type CheckoutMetadata struct {
	Source    string `json:"source,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type createCheckoutIn struct {
	Reference   string          `json:"reference"`
	AmountCents int32           `json:"amountCents"`
	ItemName    string          `json:"itemName"`
	Payer       helloasso.Payer `json:"payer"`
}

type createCheckoutOut struct {
	CheckoutIntentId int32  `json:"checkoutIntentId"`
	RedirectUrl      string `json:"redirectUrl"`
}

func (ct *Controller) CreateCheckout(c echo.Context) error {
	if !ct.checkHorizonKey(c) {
		return c.NoContent(401)
	}

	var args createCheckoutIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.createCheckout(args)
	if err != nil {
		return err
	}

	return c.JSON(201, out)
}

func (ct *Controller) createCheckout(args createCheckoutIn) (createCheckoutOut, error) {
	if args.AmountCents < 0 || args.AmountCents > 6*9000 {
		return createCheckoutOut{}, errors.New("invalid amountCents field")
	}
	if !strings.HasPrefix(args.ItemName, "Horizon") {
		return createCheckoutOut{}, errors.New("invalid itemName field: missing Horizon")
	}

	api, err := helloasso.NewApi(ct.Helloasso)
	if err != nil {
		return createCheckoutOut{}, fmt.Errorf("internal error: %s", err)
	}

	initCheckout := helloasso.InitCheckout{
		TotalAmount:      args.AmountCents,
		InitialAmount:    args.AmountCents,
		ItemName:         args.ItemName,
		BackUrl:          fmt.Sprintf(baseHorizonCallbackURL, args.Reference, "back"),
		ErrorUrl:         fmt.Sprintf(baseHorizonCallbackURL, args.Reference, "error"),
		ReturnUrl:        fmt.Sprintf(baseHorizonCallbackURL, args.Reference, "return"),
		ContainsDonation: false,
		Payer:            args.Payer,
		Metadata:         CheckoutMetadata{Source: "horizon", Reference: args.Reference},
	}
	intent, err := api.InitCheckout(initCheckout)
	if err != nil {
		// HelloAsso peut refuser des coordonnéesd invalides : on essaye à nouveau sans "payer"
		initCheckout.Payer = helloasso.Payer{}
		intent, err = api.InitCheckout(initCheckout)
	}
	if err != nil {
		return createCheckoutOut{}, err
	}

	out := createCheckoutOut{intent.Id, intent.RedirectUrl}
	return out, nil
}

func isHorizon(metadata json.RawMessage) (CheckoutMetadata, bool) {
	var m CheckoutMetadata
	if err := json.Unmarshal(metadata, &m); err != nil {
		return m, false
	}
	return m, m.Source == "horizon"
}

type checkoutOut struct {
	Reference   string `json:"reference"`
	Paid        bool   `json:"paid"`
	AmountCents int32  `json:"amountCents,omitzero"`
	OrderId     int32  `json:"orderId,omitzero"`
}

func loadCheckout(api helloasso.Api, id int32) (checkoutOut, bool, error) {
	checkout, err := api.LoadCheckout(int32(id))
	if err != nil {
		return checkoutOut{}, false, err
	}

	metadata, ok := isHorizon(checkout.Metadata)
	if !ok {
		return checkoutOut{}, false, nil
	}

	// check if the intent has been payed
	out := checkoutOut{Reference: metadata.Reference}
	if checkout.IsCompleted() {
		out.Paid = true
		out.AmountCents = checkout.Order.Amount.Total
		out.OrderId = checkout.Order.Id
	}
	return out, true, nil
}

func (ct *Controller) GetCheckout(c echo.Context) error {
	if !ct.checkHorizonKey(c) {
		return c.NoContent(401)
	}

	id, err := strconv.Atoi(c.Param("checkoutIntentId"))
	if err != nil {
		return fmt.Errorf("invalid checkoutIntentId: %s", err)
	}

	api, err := helloasso.NewApi(ct.Helloasso)
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}

	checkout, ok, err := loadCheckout(api, int32(id))
	if err != nil {
		return err
	}
	if !ok {
		return c.NoContent(404)
	}
	return c.JSON(200, checkout)
}

// HandleNotification should be called in the the main handler
func (ct *Controller) HandleNotification(c echo.Context, notif helloasso.PaymentNotification) error {
	// be defensive here, in case the main handler is lazy
	if notif.Data.Order.FormType != "Checkout" {
		return nil
	}
	// check the metadata to identify horizon notifications
	if _, ok := isHorizon(notif.Metadata); !ok {
		return nil
	}
	// now, we check the validity of the notification,
	// by retrieving the checkoutIntentId
	// Only the order seems to have the CheckoutIntentId field set
	api, err := helloasso.NewApi(ct.Helloasso)
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}
	order, err := api.LoadOrder(notif.Data.Order.Id)
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}

	out, _, err := loadCheckout(api, order.CheckoutIntentId)
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}

	// create a POST request on Horizon endpoint
	body, err := json.Marshal(out)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, horizonNotificationURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+ct.Horizon.RelaisKey)
	req.Header.Add("content-type", "application/json-patch+json")
	req.Header.Add("accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// return a status code to HelloAsso to try again later
		return c.NoContent(500)
	}

	return c.NoContent(200)
}
