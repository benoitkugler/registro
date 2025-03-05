package main

import (
	"registro/controllers/backoffice"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_backoffice.go typescript/api:../registro-web/src/clients/backoffice/logic/api.ts

func setupRoutesBackoffice(e *echo.Echo, ct *backoffice.Controller) {
	gr := e.Group("", ct.JWTMiddleware())

	// Shared

	gr.POST("/api/v1/backoffice/shared/camps", ct.GetCamps)
	gr.GET("/api/v1/backoffice/shared/personne", ct.SelectPersonne)
	gr.GET("/api/v1/backoffice/shared/structureaides", ct.GetStructureaides)

	// Onglet Camps

	gr.GET("/api/v1/backoffice/camps", ct.CampsGet)
	gr.PUT("/api/v1/backoffice/camps", ct.CampsCreate)
	gr.PUT("/api/v1/backoffice/camps-many", ct.CampsCreateMany)
	gr.POST("/api/v1/backoffice/camps", ct.CampsUpdate)
	gr.DELETE("/api/v1/backoffice/camps", ct.CampsDelete)
	gr.GET("/api/v1/backoffice/camps-taux", ct.CampsGetTaux)
	gr.POST("/api/v1/backoffice/camps-taux", ct.CampsSetTaux)

	// Onglet Inscriptions/Dossiers

	gr.GET("/api/v1/backoffice/inscriptions", ct.InscriptionsGet)
	gr.GET("/api/v1/backoffice/inscriptions/search-similaires", ct.InscriptionsSearchSimilaires)
	gr.POST("/api/v1/backoffice/inscriptions/identifie", ct.InscriptionsIdentifiePersonne)
	gr.POST("/api/v1/backoffice/inscriptions/valide", ct.InscriptionsValide)

	gr.POST("/api/v1/backoffice/dossiers/search", ct.DossiersSearch)
	gr.GET("/api/v1/backoffice/dossiers", ct.DossiersLoad)
	gr.PUT("/api/v1/backoffice/dossiers", ct.DossiersCreate)
	gr.POST("/api/v1/backoffice/dossiers", ct.DossiersUpdate)
	gr.DELETE("/api/v1/backoffice/dossiers", ct.DossiersDelete)

	gr.PUT("/api/v1/backoffice/aides", ct.AidesCreate)
	gr.POST("/api/v1/backoffice/aides", ct.AidesUpdate)
	gr.DELETE("/api/v1/backoffice/aides", ct.AidesDelete)
	gr.POST("/api/v1/backoffice/aides/justificatif", ct.AidesJustificatifUpload)
	gr.DELETE("/api/v1/backoffice/aides/justificatif", ct.AidesJustificatifDelete)

	gr.PUT("/api/v1/backoffice/participants", ct.ParticipantsCreate)
	gr.POST("/api/v1/backoffice/participants", ct.ParticipantsUpdate)
	gr.DELETE("/api/v1/backoffice/participants", ct.ParticipantsDelete)

	gr.GET("/api/v1/backoffice/paiements", ct.PaiementsCreate)
	gr.POST("/api/v1/backoffice/paiements", ct.PaiementsUpdate)
	gr.DELETE("/api/v1/backoffice/paiements", ct.PaiementsDelete)
}
