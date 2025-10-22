package main

import (
	"registro/controllers/directeurs"
	"registro/controllers/files"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro --url-only routes_misc.go typescript/api:../registro-web/src/urls.ts

func setupRoutesMisc(e *echo.Echo, dir *directeurs.Controller, fs *files.Controller) {
	// Directeurs

	e.GET("/api/v1/directeurs/equipiers/files", dir.EquipiersDownloadFiles, dir.JWTMiddlewareForQuery())

	e.POST("/api/v1/directeurs/lettre-image", dir.LettreImageUpload, dir.JWTMiddlewareForQuery())

	e.GET("/api/v1/directeurs/participants/download-liste", dir.ParticipantsDownloadListe, dir.JWTMiddlewareForQuery())

	// Documents générés

	e.GET("/api/v1/document-camp", fs.RenderDocumentCamp)

	// every endpoint expected a key=<idCrypted> query param
	e.GET("/api/v1/documents", fs.LoadDocument)
	e.GET("/api/v1/documents/miniature", fs.LoadMiniature)
}
