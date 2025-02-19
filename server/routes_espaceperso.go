package main

import (
	"registro/controllers/espaceperso"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro -http-api=/api routes_espaceperso.go typescript/api:../registro-web/src/logic/espaceperso/api.ts

func setupRoutesEspaceperso(e *echo.Echo, ct *espaceperso.Controller) {
	// client app
	e.GET("/espace-perso", ct.TmpEspaceperso)
}
