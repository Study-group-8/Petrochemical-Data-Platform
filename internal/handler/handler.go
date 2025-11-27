package handler

import (
	"net/http"
	"os"

	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/domain"
	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/service"
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

// GetAssets returns list of assets
func (h *Handler) GetAssets(c *gin.Context) {
	ctx := c.Request.Context()
	assets, err := h.assetService.GetAssets(ctx)
	if err != nil {
		h.logger.Error("GetAssets failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// GetTelemetry returns telemetry for company
func (h *Handler) GetTelemetry(c *gin.Context) {
	companyID := c.Param("company_id")
	ctx := c.Request.Context()
	start, end := service.DefaultTelemetryWindow()
	data, err := h.telemetryService.GetTelemetry(ctx, companyID, start, end)
	if err != nil {
		h.logger.Error("GetTelemetry failed", zap.Error(err), zap.String("company_id", companyID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// PostControl accepts a control command
func (h *Handler) PostControl(c *gin.Context) {
	var cmd domain.ControlCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	ctx := c.Request.Context()
	if err := h.controlService.SendControlCommand(ctx, cmd); err != nil {
		h.logger.Error("SendControlCommand failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// VerifyExportPassword is a placeholder for auth endpoint
func (h *Handler) VerifyExportPassword(c *gin.Context) {
	// Read password from request body
	var req struct {
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid payload"})
		return
	}

	// Compare with server-side secret (from env or default)
	expected := os.Getenv("ADMIN_EXPORT_PASSWORD")
	if expected == "" {
		expected = "Petrochem2025!"
	}

	if req.Password == expected {
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false})
	}
}

// HandleWebSocket placeholder
func (h *Handler) HandleWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "websocket not implemented"})
}
