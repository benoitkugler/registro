package main

import (
	"registro/controllers/equipier"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_equipier.go typescript/api:../registro-web/src/clients/equipier/logic/api.ts

func setupRoutesEquipier(e *echo.Echo, ct *equipier.Controller) {
	// JSON API
	e.GET("/api/v1/equipier", ct.Load)
	e.GET("/api/v1/equipier/joomeo", ct.LoadJoomeo)
	e.POST("/api/v1/equipier", ct.Update)
	e.POST("/api/v1/equipier/charte", ct.UpdateCharte)
	e.PUT("/api/v1/equipier/upload", ct.UploadDocument)
	e.DELETE("/api/v1/equipier/upload", ct.DeleteDocument)
}
