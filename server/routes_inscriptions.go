package main

import (
	"registro/controllers/inscriptions"
	"registro/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_inscriptions.go typescript/api:../registro-web/src/logic/inscriptions/api.ts

func setupRoutesInscriptions(e *echo.Echo, ct *inscriptions.Controller) {
	// client app
	e.GET(inscriptions.EndpointInscription, func(c echo.Context) error {
		return c.File("static/inscription/index.html")
	}, middleware.Gzip(), utils.NoCache)
	e.GET(inscriptions.EndpointConfirmeInscription, ct.ConfirmeInscription)

	// JSON API
	e.GET("/inscription/v1/load", ct.LoadData)
	e.PUT("/inscription/v1/save", ct.SaveInscription)
	e.GET("/inscription/v1/search", ct.SearchHistory)
}
