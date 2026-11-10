package main

import (
	"registro/controllers/compta"

	"github.com/labstack/echo/v4"
)

//go:generate ../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro routes_compta.go typescript/api:../registro-web/src/clients/compta/logic/api.ts

func setupRoutesCompta(e *echo.Echo, ct *compta.Controller) {
	// no token yet for the loggin route
	e.GET("/api/v1/compta/loggin", ct.Loggin)

	gr := e.Group("", ct.JWTMiddleware())

	gr.GET("/api/v1/compta/camps", ct.GetCamps)
	gr.GET("/api/v1/compta/camp", ct.GetCompta)
	gr.PUT("/api/v1/compta/camp", ct.CreateCompta)
	gr.POST("/api/v1/compta/camp", ct.UpdateCompta)
	gr.DELETE("/api/v1/compta/camp", ct.DeleteCompta)

	gr.POST("/api/v1/compta/camp/state", ct.SetComptaState)
	gr.GET("/api/v1/compta/camp/details", ct.GetCampDetails)

	gr.GET("/api/v1/compta/configs", ct.GetComptaConfigs)
	gr.PUT("/api/v1/compta/configs", ct.CreateComptaConfig)
	gr.POST("/api/v1/compta/configs", ct.UpdateComptaConfig)
	gr.DELETE("/api/v1/compta/configs", ct.DeleteComptaConfig)

	gr.PUT("/api/v1/compta/budget", ct.CreateBudget)
	gr.GET("/api/v1/compta/budget", ct.GetEstimatedBudget)
	gr.POST("/api/v1/compta/budget", ct.UpdateEstimatedBudget)
	gr.GET("/api/v1/compta/budget/previous-budgets", ct.GetPreviousBudgets)
	gr.GET("/api/v1/compta/budget/activities", ct.GetBudgetActivities)

	gr.GET("/api/v1/compta/frais-mdr", ct.FraisMDRLoad)
	gr.POST("/api/v1/compta/frais-mdr", ct.FraisMDRUpdate)
	gr.GET("/api/v1/compta/frais-mdr/bus-expense", ct.FraisMDRGetBusExpense)
	gr.POST("/api/v1/compta/frais-mdr/bus-expense", ct.FraisMDRSetBusExpense)

	gr.GET("/api/v1/compta/records", ct.GetRecords)
	gr.PUT("/api/v1/compta/records", ct.CreateRecord)
	gr.POST("/api/v1/compta/records", ct.UpdateRecord)
	gr.PUT("/api/v1/compta/records/subrecord", ct.AddSubrecord)
	gr.DELETE("/api/v1/compta/records/subrecord", ct.DeleteSubrecord)
	gr.POST("/api/v1/compta/records/merge", ct.MergeRecordTo)
	gr.POST("/api/v1/compta/records/review", ct.AddRecordReview)
	gr.POST("/api/v1/compta/records/upload", ct.UploadAndCreate)

	gr.POST("/api/v1/compta/records/justificatif", ct.UploadJustificatif)
	gr.DELETE("/api/v1/compta/records/justificatif", ct.RemoveJustificatif)

	gr.POST("/api/v1/compta/download-bilan-camps", ct.DownloadBilanCamps)

	// File URLs are opened directly by the client and take their token in the query.
	e.GET("/api/v1/compta/records/justificatif", ct.DownloadJustificatif, ct.JWTMiddlewareForQuery())             // url-only
	e.GET("/api/v1/compta/records/miniature", ct.GetMiniature, ct.JWTMiddlewareForQuery())                        // url-only
	e.GET("/api/v1/compta/download-ecritures", ct.DownloadEcritures, ct.JWTMiddlewareForQuery())                  // url-only
	e.GET("/api/v1/compta/records/download", ct.DownloadRecords, ct.JWTMiddlewareForQuery())                      // url-only
	e.GET("/api/v1/compta/records/download-stream", ct.DownloadRecordsAndFilesStream, ct.JWTMiddlewareForQuery()) // url-only
}
