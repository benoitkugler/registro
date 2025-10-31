package main

import (
	"registro/controllers/files"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro --url-only routes_misc.go typescript/api:../registro-web/src/urls.ts

func setupRoutesMisc(e *echo.Echo, fs *files.Controller) {
	// Directeurs

	// Documents générés

	e.GET("/api/v1/document-camp", fs.RenderDocumentCamp)

	// every endpoint expected a key=<idCrypted> query param
	e.GET("/api/v1/documents", fs.LoadDocument)
	e.GET("/api/v1/documents/miniature", fs.LoadMiniature)
}
