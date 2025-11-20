package main

import (
	"registro/controllers/espaceperso"
	"registro/controllers/services"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_services.go typescript/api:../registro-web/src/clients/services/logic/api.ts

func setupRoutesServices(e *echo.Echo, ctServices *services.Controller, ctEspaceperso *espaceperso.Controller) {
	// validation partage fiche sanitaire
	e.POST("/api/v1/espaceperso/fichesanitaires/transfert", ctEspaceperso.ValideTransfertFicheSanitaire)

	e.GET("/api/v1/services/search-mail", ctServices.SearchMail)
}
