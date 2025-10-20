package handler

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name        string    `json:"name" example:"ShortLink Service"`
	Version     string    `json:"version" example:"1.0.0"`
	Description string    `json:"description" example:"URL shortening service with analytics"`
	StartTime   time.Time `json:"start_time" example:"2024-01-15T10:30:45Z"`
	Uptime      string    `json:"uptime" example:"72h30m15s"`
	GoVersion   string    `json:"go_version" example:"go1.21.0"`
}

// HealthStatus represents health check status
type HealthStatus struct {
	Status    string    `json:"status" example:"healthy"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:45Z"`
	Uptime    string    `json:"uptime" example:"72h30m15s"`
	Version   string    `json:"version" example:"1.0.0"`
}

// GetServiceInfo godoc
// @Summary Get service information
// @Description Returns basic information about the service
// @Tags info
// @Produce json
// @Success 200 {object} ServiceInfo
// @Router / [get]
func (h *HealthHandler) GetServiceInfo(c *gin.Context) {
	uptime := time.Since(h.startTime)

	info := ServiceInfo{
		Name:        "ShortLink Service",
		Version:     h.version,
		Description: "High-performance URL shortening service with analytics tracking, QR code generation, and JWT/API key authentication",
		StartTime:   h.startTime,
		Uptime:      uptime.String(),
		GoVersion:   runtime.Version(),
	}

	RespondOK(c, info)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(h.startTime)

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   h.version,
	}

	RespondOK(c, status)
}

func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
	RespondOK(c, gin.H{
		"status":    "ready",
		"timestamp": time.Now(),
	})
}

func (h *HealthHandler) LivenessProbe(c *gin.Context) {
	RespondOK(c, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}
