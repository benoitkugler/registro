package main

import (
	"registro/controllers/espaceperso"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_services.go typescript/api:../registro-web/src/clients/services/logic/api.ts

func setupRoutesServices(e *echo.Echo, ct *espaceperso.Controller) {
	// validation partage fiche sanitaire
	e.POST("/api/v1/espaceperso/fichesanitaires/transfert", ct.ValideTransfertFicheSanitaire)
}
