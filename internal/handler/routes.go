package handler

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		api.GET("/assets", handler.GetAssets)
		api.GET("/telemetry/:company_id", handler.GetTelemetry)
		api.POST("/control", handler.PostControl)
		api.POST("/auth/verify-export-password", handler.VerifyExportPassword)
	}

	// WebSocket endpoint
	r.GET("/ws", handler.HandleWebSocket)
}
