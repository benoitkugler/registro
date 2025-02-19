package main

import (
	"registro/controllers/backoffice"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_backoffice.go typescript/api:../registro-web/src/clients/backoffice/logic/api.ts

func setupRoutesBackoffice(e *echo.Echo, ct *backoffice.Controller) {
	gr := e.Group("", ct.JWTMiddleware())
	gr.GET("/api/v1/backoffice/camps", ct.CampsGet)
	gr.PUT("/api/v1/backoffice/camps", ct.CampsCreate)
	gr.PUT("/api/v1/backoffice/camps-many", ct.CampsCreateMany)
	gr.POST("/api/v1/backoffice/camps", ct.CampsUpdate)
	gr.DELETE("/api/v1/backoffice/camps", ct.CampsDelete)
	gr.GET("/api/v1/backoffice/camps-taux", ct.CampsGetTaux)
	gr.POST("/api/v1/backoffice/camps-taux", ct.CampsSetTaux)

	gr.GET("/api/v1/backoffice/inscriptions", ct.InscriptionsGet)
}
