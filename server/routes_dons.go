package main

import (
	"registro/controllers/dons"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_dons.go typescript/api:../registro-web/src/clients/dons/logic/api.ts

func setupRoutesDons(e *echo.Echo, ct *dons.Controller) {
	// no token yet for the loggin route
	e.GET("/api/v1/loggin", ct.Loggin)

	gr := e.Group("", ct.JWTMiddleware())

	gr.GET("/api/v1/dons", ct.LoadDons)
	gr.POST("/api/v1/dons", ct.UpdateDon)
	gr.PUT("/api/v1/dons", ct.CreateDon)
	gr.DELETE("/api/v1/dons", ct.DeleteDon)

	gr.GET("/api/v1/dons/search-personnes", ct.SearchPersonnes)
	gr.GET("/api/v1/dons/search-organismes", ct.SearchOrganismes)

	e.GET("/api/v1/dons/download-recus-fiscaux", ct.DownloadRecusFiscaux, ct.JWTMiddlewareForQuery()) // url-only
}
