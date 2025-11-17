package main

import (
	"registro/controllers/backoffice"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_backoffice.go typescript/api:../registro-web/src/clients/backoffice/logic/api.ts

func setupRoutesBackoffice(e *echo.Echo, ct *backoffice.Controller) {
	// no token yet for the loggin route
	e.GET("/api/v1/backoffice/loggin", ct.Loggin)

	gr := e.Group("", ct.JWTMiddleware())

	// Shared

	gr.POST("/api/v1/backoffice/shared/camps", ct.GetCamps)
	gr.GET("/api/v1/backoffice/shared/personne", ct.SelectPersonne)

	// Onglet Camps

	gr.GET("/api/v1/backoffice/camps", ct.CampsGet)
	gr.PUT("/api/v1/backoffice/camps", ct.CampsCreate)
	gr.PUT("/api/v1/backoffice/camps-many", ct.CampsCreateMany)
	gr.POST("/api/v1/backoffice/camps", ct.CampsUpdate)
	gr.DELETE("/api/v1/backoffice/camps", ct.CampsDelete)
	gr.GET("/api/v1/backoffice/camps-taux", ct.CampsGetTaux)
	gr.POST("/api/v1/backoffice/camps-taux", ct.CampsSetTaux)

	gr.POST("/api/v1/backoffice/camps/open", ct.CampsOuvreInscriptions)

	gr.GET("/api/v1/backoffice/camps/load", ct.CampsLoad)
	gr.GET("/api/v1/backoffice/camps/documents", ct.CampsDocuments)

	gr.PUT("/api/v1/backoffice/participants", ct.ParticipantsCreate)
	gr.POST("/api/v1/backoffice/participants", ct.ParticipantsUpdate)
	gr.DELETE("/api/v1/backoffice/participants", ct.ParticipantsDelete)

	gr.POST("/api/v1/backoffice/participants/move", ct.ParticipantsMove)
	gr.POST("/api/v1/backoffice/participants/place-liberee", ct.ParticipantsSetPlaceLiberee)

	gr.GET("/api/v1/backoffice/camps/joomeo", ct.CampsLoadAlbums)
	gr.PUT("/api/v1/backoffice/camps/joomeo", ct.CampsCreateAlbums)
	gr.POST("/api/v1/backoffice/camps/joomeo", ct.CampsAddDirecteursToAlbums)

	e.GET("/api/v1/backoffice/camps/download-participants", ct.CampsDownloadParticipants, ct.JWTMiddlewareForQuery()) // url-only

	gr.PUT("/api/v1/backoffice/camps/equipiers", ct.CampsCreateEquipier)

	// Onglet Inscriptions/Dossiers

	gr.GET("/api/v1/backoffice/inscriptions", ct.InscriptionsGet)
	gr.GET("/api/v1/backoffice/inscriptions/search-similaires", ct.InscriptionsSearchSimilaires)
	gr.POST("/api/v1/backoffice/inscriptions/identifie", ct.InscriptionsIdentifiePersonne)
	gr.POST("/api/v1/backoffice/inscriptions/valide/hint", ct.InscriptionsHintValide)
	gr.POST("/api/v1/backoffice/inscriptions/valide", ct.InscriptionsValide)

	gr.POST("/api/v1/backoffice/dossiers/search", ct.DossiersSearch)
	gr.GET("/api/v1/backoffice/dossiers", ct.DossiersLoad)
	gr.PUT("/api/v1/backoffice/dossiers", ct.DossiersCreate)
	gr.POST("/api/v1/backoffice/dossiers", ct.DossiersUpdate)
	gr.DELETE("/api/v1/backoffice/dossiers", ct.DossiersDelete)

	gr.PUT("/api/v1/backoffice/dossiers/remises-hints", ct.DossiersRemisesHint)
	gr.POST("/api/v1/backoffice/dossiers/remises-hints", ct.DossiersApplyRemisesHints)

	gr.POST("/api/v1/backoffice/dossiers/merge", ct.DossiersMerge)

	gr.GET("/api/v1/backoffice/structureaides", ct.StructureaidesGet)
	gr.PUT("/api/v1/backoffice/structureaides", ct.StructureaideCreate)
	gr.POST("/api/v1/backoffice/structureaides", ct.StructureaideUpdate)
	gr.DELETE("/api/v1/backoffice/structureaides", ct.StructureaideDelete)

	gr.PUT("/api/v1/backoffice/aides", ct.AidesCreate)
	gr.POST("/api/v1/backoffice/aides", ct.AidesUpdate)
	gr.DELETE("/api/v1/backoffice/aides", ct.AidesDelete)
	gr.POST("/api/v1/backoffice/aides/justificatif", ct.AidesJustificatifUpload)
	gr.DELETE("/api/v1/backoffice/aides/justificatif", ct.AidesJustificatifDelete)

	gr.GET("/api/v1/backoffice/paiements", ct.PaiementsCreate)
	gr.POST("/api/v1/backoffice/paiements", ct.PaiementsUpdate)
	gr.DELETE("/api/v1/backoffice/paiements", ct.PaiementsDelete)

	gr.POST("/api/v1/backoffice/events/message", ct.EventsSendMessage)
	gr.DELETE("/api/v1/backoffice/events", ct.EventsDelete)
	gr.POST("/api/v1/backoffice/events/message/seen", ct.EventsMarkMessagesSeen)
	gr.POST("/api/v1/backoffice/events/facture", ct.EventsSendFacture)
	gr.GET("/api/v1/backoffice/events/documents-camp", ct.EventsSendDocumentsCampPreview)
	gr.POST("/api/v1/backoffice/events/documents-camp", ct.EventsSendDocumentsCamp)
	gr.POST("/api/v1/backoffice/events/sondage", ct.EventsSendSondages)
	gr.GET("/api/v1/backoffice/events/relance-paiement", ct.EventsSendRelancePaiementPreview)
	gr.POST("/api/v1/backoffice/events/relance-paiement", ct.EventsSendRelancePaiement)

	// Onglet Annuaire

	gr.GET("/api/v1/backoffice/personnes/search", ct.PersonnesGet)
	gr.GET("/api/v1/backoffice/personnes", ct.PersonnesLoad)
	gr.PUT("/api/v1/backoffice/personnes", ct.PersonnesCreate)
	gr.POST("/api/v1/backoffice/personnes", ct.PersonnesUpdate)
}
