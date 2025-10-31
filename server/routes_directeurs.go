package main

import (
	"registro/controllers/directeurs"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_directeurs.go typescript/api:../registro-web/src/clients/directeurs/logic/api.ts

func setupRoutesDirecteurs(e *echo.Echo, ct *directeurs.Controller) {
	// no token yet for the loggin route
	e.POST("/api/v1/directeurs/shared/camps", ct.GetCamps)
	e.GET("/api/v1/directeurs/loggin", ct.Loggin)

	// public image service, secured by key
	e.GET(directeurs.EndpointLettreImages, ct.LettreImageGet) // ignore

	// file download URLs (see also routes_misc.go)
	e.GET("/api/v1/directeurs/documents/stream-files", ct.DocumentsStreamFiles, ct.JWTMiddlewareForQuery())                            // url-only
	e.GET("/api/v1/directeurs/documents/download-fiches-sanitaires", ct.DocumentsDownloadFichesSanitaires, ct.JWTMiddlewareForQuery()) // url-only
	e.GET("/api/v1/directeurs/participants/download-liste", ct.ParticipantsDownloadListe, ct.JWTMiddlewareForQuery())                  // url-only
	e.GET("/api/v1/directeurs/equipiers/files", ct.EquipiersDownloadFiles, ct.JWTMiddlewareForQuery())                                 // url-only
	e.POST("/api/v1/directeurs/lettre-image", ct.LettreImageUpload, ct.JWTMiddlewareForQuery())                                        // url-only

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

	gr.GET("/api/v1/directeurs/participants/files", ct.ParticipantsLoadFiles)
	gr.POST("/api/v1/directeurs/participants/relance-documents", ct.ParticipantsRelanceDocuments)

	// Messages
	gr.GET("/api/v1/directeurs/participants/messages", ct.ParticipantsMessagesLoad)
	gr.PUT("/api/v1/directeurs/participants/messages", ct.ParticipantsMessagesCreate)
	gr.POST("/api/v1/directeurs/participants/messages/seen", ct.ParticipantsMessageSetSeen)

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

	// Joomeo
	gr.GET("/api/v1/directeurs/joomeo", ct.JoomeoLoad)
	gr.PUT("/api/v1/directeurs/joomeo", ct.JoomeoInvite)
	gr.POST("/api/v1/directeurs/joomeo", ct.JoomeoSetUploader)
	gr.DELETE("/api/v1/directeurs/joomeo", ct.JoomeoUnlinkContact)
}
