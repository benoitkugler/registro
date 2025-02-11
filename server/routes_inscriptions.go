package main

import (
	"registro/controllers/inscriptions"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_inscriptions.go typescript/api:../registro-web/src/logic/inscriptions/api.ts

func setupRoutesInscriptions(e *echo.Echo, ct *inscriptions.Controller) {
	e.GET("/inscription/v1/load", ct.LoadData)
	e.PUT("/inscription/v1/save", ct.SaveInscription)
	e.GET("/inscription/v1/search", ct.SearchHistory)
	e.GET(inscriptions.EndpointConfirmeInscription, ct.ConfirmeInscription)
}
