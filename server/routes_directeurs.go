package main

import (
	"registro/controllers/directeurs"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro -http-api=/api routes_directeurs.go typescript/api:../registro-web/src/clients/directeurs/logic/api.ts

func setupRoutesDirecteurs(e *echo.Echo, ct *directeurs.Controller) {
	// no token yet for the loggin route
	e.POST("/api/v1/directeurs/shared/camps", ct.GetCamps)
	e.GET("/api/v1/directeurs/loggin", ct.Loggin)

	e.GET("/api/v1/directeurs/equipiers/files", ct.EquipiersDownloadFiles, ct.JWTMiddlewareForQuery())
	e.GET("/api/v1/directeurs/participants/stream-fiches-sanitaires", ct.ParticipantsStreamFichesAndVaccins, ct.JWTMiddlewareForQuery())

	// public image service, secured by key
	e.GET(directeurs.EndpointLettreImages, ct.LettreImageGet)
	e.POST("/api/v1/directeurs/lettre-image", ct.LettreImageUpload, ct.JWTMiddlewareForQuery())

	e.GET("/service/directeurs/vetements", ct.VetementsRender, ct.JWTMiddlewareForQuery())

	gr := e.Group("", ct.JWTMiddleware())

	// Shared

	gr.GET("/api/v1/directeurs/shared/personne", ct.SelectPersonne)

	// Inscriptions

	gr.GET("/api/v1/directeurs/inscriptions", ct.InscriptionsGet)
	gr.GET("/api/v1/directeurs/inscriptions/search-similaires", ct.InscriptionsSearchSimilaires)
	gr.POST("/api/v1/directeurs/inscriptions/identifie", ct.InscriptionsIdentifiePersonne)
	gr.POST("/api/v1/directeurs/inscriptions/valide/hint", ct.InscriptionsHintValide)
	gr.POST("/api/v1/directeurs/inscriptions/valide", ct.InscriptionsValide)

	// Participants
	gr.GET("/api/v1/directeurs/participants", ct.ParticipantsGet)
	gr.POST("/api/v1/directeurs/participants", ct.ParticipantsUpdate)
	gr.GET("/api/v1/directeurs/participants/fiches-sanitaires", ct.ParticipantsGetFichesSanitaires)
	gr.GET("/api/v1/directeurs/participants/download-fiche-sanitaire", ct.ParticipantsDownloadFicheSanitaire)
	gr.GET("/api/v1/directeurs/participants/download-fiches-sanitaires", ct.ParticipantsDownloadAllFichesSanitaires)

	// Equipiers
	gr.GET("/api/v1/directeurs/equipiers", ct.EquipiersGet)
	gr.PUT("/api/v1/directeurs/equipiers", ct.EquipiersCreate)
	gr.DELETE("/api/v1/directeurs/equipiers", ct.EquipiersDelete)
	gr.POST("/api/v1/directeurs/equipiers/invite", ct.EquipiersInvite)
	gr.GET("/api/v1/directeurs/equipiers/demandes", ct.EquipiersDemandesGet)
	gr.POST("/api/v1/directeurs/equipiers/demandes", ct.EquipiersDemandeSet)

	// Lettre
	gr.GET("/api/v1/directeurs/lettre", ct.LettreGet)
	gr.POST("/api/v1/directeurs/lettre", ct.LettreUpdate)

	// Vetements
	gr.GET("/api/v1/directeurs/vetements", ct.VetementsGet)
	gr.POST("/api/v1/directeurs/vetements", ct.VetementsUpdate)

	// Documents
	gr.GET("/api/v1/directeurs/documents", ct.DocumentsGet)
	gr.POST("/api/v1/directeurs/documents", ct.DocumentsUpdateToShow)
	gr.POST("/api/v1/directeurs/documents/to-download", ct.DocumentsUploadToDownload)
	gr.DELETE("/api/v1/directeurs/documents/to-download", ct.DocumentsDeleteToDownload)
	gr.PUT("/api/v1/directeurs/documents/demande", ct.DocumentsCreateDemande)
	gr.POST("/api/v1/directeurs/documents/demande", ct.DocumentsUpdateDemande)
	gr.DELETE("/api/v1/directeurs/documents/demande", ct.DocumentsDeleteDemande)
	gr.POST("/api/v1/directeurs/documents/demande/file", ct.DocumentsUploadDemandeFile)
	gr.DELETE("/api/v1/directeurs/documents/demande/file", ct.DocumentsDeleteDemandeFile)
	gr.POST("/api/v1/directeurs/documents/demande/apply", ct.DocumentsApplyDemande)
	gr.DELETE("/api/v1/directeurs/documents/demande/apply", ct.DocumentsUnapplyDemande)
}
