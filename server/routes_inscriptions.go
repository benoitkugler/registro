package main

import (
	"registro/controllers/inscriptions"
	"registro/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro -http-api=/api routes_inscriptions.go typescript/api:../registro-web/src/clients/inscriptions/logic/api.ts

func setupRoutesInscriptions(e *echo.Echo, ct *inscriptions.Controller) {
	// client app
	e.GET(inscriptions.EndpointInscription, func(c echo.Context) error {
		return c.File("static/inscription/index.html")
	}, middleware.Gzip(), utils.NoCache)
	e.GET(inscriptions.EndpointConfirmeInscription, ct.ConfirmeInscription)

	// JSON API
	e.GET("/api/v1/inscription/load", ct.LoadData)
	e.PUT("/api/v1/inscription/save", ct.SaveInscription)
	e.GET("/api/v1/inscription/search", ct.SearchHistory)
}
