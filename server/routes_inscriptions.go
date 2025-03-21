package main

import (
	"registro/controllers/inscriptions"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro -http-api=/api routes_inscriptions.go typescript/api:../registro-web/src/clients/inscription/logic/api.ts

func setupRoutesInscriptions(e *echo.Echo, ct *inscriptions.Controller) {
	e.GET(inscriptions.EndpointConfirmeInscription, ct.ConfirmeInscription)

	// JSON API
	e.GET("/api/v1/inscription/load", ct.LoadData)
	e.PUT("/api/v1/inscription/save", ct.SaveInscription)
	e.GET("/api/v1/inscription/search", ct.SearchHistory)
}
