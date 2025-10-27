package main

import (
	"registro/controllers/espaceperso"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_espaceperso.go typescript/api:../registro-web/src/clients/espaceperso/logic/api.ts

func setupRoutesEspaceperso(e *echo.Echo, ct *espaceperso.Controller) {
	// Download endpoints
	e.GET("/api/v1/espaceperso/download/attestation", ct.DownloadAttestationPresence) // url-only
	e.GET("/api/v1/espaceperso/download/facture", ct.DownloadFacture)                 // url-only

	// JSON API
	e.GET("/api/v1/espaceperso", ct.Load)
	e.POST("/api/v1/espaceperso/message", ct.SendMessage)
	e.POST("/api/v1/espaceperso/participants", ct.UpdateParticipants)
	e.POST("/api/v1/espaceperso/aide", ct.CreateAide)
	e.GET("/api/v1/espaceperso/structureaides", ct.GetStructureaides)

	e.GET("/api/v1/espaceperso/joomeo", ct.LoadJoomeo)

	e.POST("/api/v1/espaceperso/events/accept-place-liberee", ct.AcceptePlaceLiberee)

	// Sondages
	e.GET("/api/v1/espaceperso/sondages", ct.LoadSondages)
	e.POST("/api/v1/espaceperso/sondages", ct.UpdateSondages)

	// Documents
	e.GET("/api/v1/espaceperso/documents", ct.LoadDocuments)
	e.POST("/api/v1/espaceperso/documents", ct.UploadDocument)
	e.DELETE("/api/v1/espaceperso/documents", ct.DeleteDocument)
	e.POST("/api/v1/espaceperso/documents/charte", ct.AccepteCharte)

	e.POST("/api/v1/espaceperso/fichesanitaires", ct.UpdateFichesanitaire)
	e.PUT("/api/v1/espaceperso/fichesanitaires/transfert", ct.TransfertFicheSanitaire)
}
