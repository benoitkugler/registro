package main

import (
	"registro/controllers/central"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes.go typescript/api:../clients/central/api.ts

func setupRoutesCentral(e *echo.Echo, ct *central.Controller) {
	gr := e.Group("/api/v1/central", ct.JWTMiddleware())

	gr.GET("/camps", ct.CampsGet)
	gr.PUT("/camps", ct.CampsCreate)
	gr.POST("/camps", ct.CampsUpdate)
	gr.DELETE("/camps", ct.CampsDelete)
}
