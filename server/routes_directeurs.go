package main

import (
	"registro/controllers/directeurs"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_directeurs.go typescript/api:../registro-web/src/clients/directeurs/logic/api.ts

func setupRoutesDirecteurs(e *echo.Echo, ct *directeurs.Controller) {
	// no token yet for the loggin route
	e.GET("/api/v1/directeurs/loggin", ct.Loggin)

	gr := e.Group("", ct.JWTMiddleware())

	// Shared

	gr.POST("/api/v1/directeurs/shared/camps", ct.GetCamps)
	gr.GET("/api/v1/directeurs/shared/personne", ct.SelectPersonne)

	// Inscriptions

	gr.GET("/api/v1/directeurs/inscriptions", ct.InscriptionsGet)
	gr.GET("/api/v1/directeurs/inscriptions/search-similaires", ct.InscriptionsSearchSimilaires)
	gr.POST("/api/v1/directeurs/inscriptions/identifie", ct.InscriptionsIdentifiePersonne)
	gr.POST("/api/v1/directeurs/inscriptions/valide", ct.InscriptionsValide)

	// Participants
	gr.GET("/api/v1/directeurs/participants", ct.ParticipantsGet)
	gr.POST("/api/v1/directeurs/participants", ct.ParticipantsUpdate)
}
