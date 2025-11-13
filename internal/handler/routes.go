package handler

import (
	"petrochemical-data-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	assetService     *service.AssetService
	telemetryService *service.TelemetryService
	controlService   *service.ControlService
	logger           *zap.Logger
}

func NewHandler(assetSvc *service.AssetService, telemetrySvc *service.TelemetryService, controlSvc *service.ControlService, logger *zap.Logger) *Handler {
	return &Handler{
		assetService:     assetSvc,
		telemetryService: telemetrySvc,
		controlService:   controlSvc,
		logger:           logger,
	}
}

func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		api.GET("/assets", handler.GetAssets)
		api.GET("/telemetry/:sensor_id", handler.GetTelemetry)
		api.POST("/control", handler.PostControl)
		api.POST("/auth/verify-export-password", handler.VerifyExportPassword)
	}

	// WebSocket endpoint
	r.GET("/ws", handler.HandleWebSocket)
}
