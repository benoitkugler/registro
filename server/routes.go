package main

import (
	"registro/controllers/central"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes.go typescript/api:../registro-web/src/logic/app/api.ts

func setupRoutesCentral(e *echo.Echo, ct *central.Controller) {
	gr := e.Group("", ct.JWTMiddleware())
	gr.GET("/api/v1/app/camps", ct.CampsGet)
	gr.PUT("/api/v1/app/camps", ct.CampsCreate)
	gr.PUT("/api/v1/app/camps-many", ct.CampsCreateMany)
	gr.POST("/api/v1/app/camps", ct.CampsUpdate)
	gr.DELETE("/api/v1/app/camps", ct.CampsDelete)
	gr.GET("/api/v1/app/camps-taux", ct.CampsGetTaux)
	gr.POST("/api/v1/app/camps-taux", ct.CampsSetTaux)
}
