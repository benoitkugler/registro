package main

import (
	"registro/controllers/espaceperso"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro -http-api=/api routes_espaceperso.go typescript/api:../registro-web/src/clients/espaceperso/logic/api.ts

func setupRoutesEspaceperso(e *echo.Echo, ct *espaceperso.Controller) {
	// JSON API
	e.GET("/api/v1/espaceperso", ct.Load)
	e.POST("/api/v1/espaceperso/message", ct.SendMessage)
	e.POST("/api/v1/espaceperso/participants", ct.UpdateParticipants)
	e.POST("/api/v1/espaceperso/aide", ct.CreateAide)
}
