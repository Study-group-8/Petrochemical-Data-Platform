package handler

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strconv"
	"time"

	"petrochemical-data-platform/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetAssets handles GET /api/v1/assets
func (h *Handler) GetAssets(c *gin.Context) {
	assets, err := h.assetService.GetAssets(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve assets"})
		return
	}

	c.JSON(http.StatusOK, assets)
}

// GetTelemetry handles GET /api/v1/telemetry/{company_id}
func (h *Handler) GetTelemetry(c *gin.Context) {
	companyID := c.Param("company_id")

	// Parse query parameters for time range
	startStr := c.DefaultQuery("start", "")
	endStr := c.DefaultQuery("end", "")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
			return
		}
	} else {
		start = time.Now().Add(-time.Hour) // Default to last hour
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
			return
		}
	} else {
		end = time.Now()
	}

	data, err := h.telemetryService.GetTelemetry(c.Request.Context(), companyID, start, end)
	if err != nil {
		h.logger.Error("Failed to get telemetry", zap.Error(err), zap.String("company_id", companyID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve telemetry data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"company_id": companyID,
		"data":       data,
	})
}

// PostControl handles POST /api/v1/control
func (h *Handler) PostControl(c *gin.Context) {
	var cmd domain.ControlCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	cmd.ID = strconv.FormatInt(time.Now().UnixNano(), 10) // Simple ID generation
	cmd.Status = "pending"
	cmd.CreatedAt = time.Now()

	if err := h.controlService.SendControlCommand(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to send control command", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send control command"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Control command sent",
		"command_id": cmd.ID,
	})
}

// HandleWebSocket handles WebSocket connections for real-time data
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Placeholder for WebSocket implementation
	// In production, this would upgrade to WebSocket and stream real-time data
	c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not implemented yet"})
}

// VerifyExportPassword handles POST /api/v1/auth/verify-export-password
func (h *Handler) VerifyExportPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Get password from environment variable
	adminPassword := os.Getenv("ADMIN_EXPORT_PASSWORD")
	if adminPassword == "" {
		h.logger.Error("ADMIN_EXPORT_PASSWORD environment variable not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(adminPassword)) == 1 {
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid password"})
	}
}
