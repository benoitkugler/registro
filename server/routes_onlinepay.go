package main

import (
	"registro/controllers/espaceperso"

	"github.com/labstack/echo/v4"
)

// setupRoutesOnlinePay registers provider-facing routes. It lives outside
// routes_espaceperso.go on purpose: the webhook must not be generated in
// the family's TypeScript API client.
func setupRoutesOnlinePay(e *echo.Echo, ct *espaceperso.Controller) {
	e.POST("/api/v1/onlinepay/webhook", ct.HandleOnlinePaymentNotification)
}
