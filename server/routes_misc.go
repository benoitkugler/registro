package main

import (
	"registro/controllers/directeurs"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro --url-only routes_misc.go typescript/api:../registro-web/src/urls.ts

func setupRoutesMisc(e *echo.Echo, dir *directeurs.Controller) {
	// Directeurs

	e.GET("/api/v1/directeurs/equipiers/files", dir.EquipiersDownloadFiles, dir.JWTMiddlewareForQuery())
	e.GET("/api/v1/directeurs/participants/stream-fiches-sanitaires", dir.ParticipantsStreamFichesAndVaccins, dir.JWTMiddlewareForQuery())
	e.GET("/api/v1/directeurs/documents/stream-documents", dir.DocumentsStreamUploaded, dir.JWTMiddlewareForQuery())

	e.POST("/api/v1/directeurs/lettre-image", dir.LettreImageUpload, dir.JWTMiddlewareForQuery())

	e.GET("/service/directeurs/vetements", dir.VetementsRender, dir.JWTMiddlewareForQuery())
}
