package handler

import (
	"context"
	"errors"
	"net/http"

	"service-short-link/internal/domain"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkUseCase domain.LinkUseCase
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(linkUseCase domain.LinkUseCase) *LinkHandler {
	return &LinkHandler{
		linkUseCase: linkUseCase,
	}
}

// CreateLink godoc
// @Summary Create a new short link
// @Description Creates a new short link from a long URL. If short_code is provided, it must follow validation rules. Expiry is automatically set to default from config.
// @Tags links
// @Accept json
// @Produce json
// @Param request body domain.CreateLinkRequest true "Create link request - original_url is required, short_code is optional but must be 5-10 alphanumeric characters. Platform/role/channel_code extracted from JWT for internal services."
// @Success 201 {object} handler.SuccessEnvelope{data=domain.CreateLinkResponse} "Link created successfully"
// @Failure 400 {object} handler.ErrorEnvelope "Bad request - invalid input or short code already exists"
// @Failure 401 {object} handler.ErrorEnvelope "Unauthorized - valid JWT token or API key required"
// @Failure 500 {object} handler.ErrorEnvelope "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /api/v1/links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req domain.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, CodeBadRequest, "Invalid request body", err.Error())
		return
	}

	var platform, role, channelCode *string
	if isInternal, exists := c.Get("auth_is_internal"); exists {
		if internal, ok := isInternal.(bool); ok && internal {
			if p, exists := c.Get("auth_platform"); exists {
				if pStr, ok := p.(string); ok && pStr != "" {
					platform = &pStr
				}
			}
			if r, exists := c.Get("auth_role"); exists {
				if rStr, ok := r.(string); ok && rStr != "" {
					role = &rStr
				}
			}
			if cc, exists := c.Get("auth_channel_code"); exists {
				if ccStr, ok := cc.(string); ok && ccStr != "" {
					channelCode = &ccStr
				}
			}
		}
	}

	response, err := h.linkUseCase.CreateLink(&req, platform, role, channelCode)
	if err != nil {
		switch err {
		case domain.ErrInvalidURL:
			RespondError(c, http.StatusBadRequest, CodeBadRequest, "Invalid URL format", nil)
		case domain.ErrShortCodeExists:
			RespondError(c, http.StatusConflict, CodeConflict, "Short code already exists", nil)
		case domain.ErrInvalidShortCode:
			RespondError(c, http.StatusBadRequest, CodeBadRequest, "Invalid short code format - must be 5-10 alphanumeric characters and not a reserved word", nil)
		default:
			RespondError(c, http.StatusInternalServerError, CodeInternalError, "An error occurred", nil)
		}
		return
	}

	RespondCreated(c, response)
}

// RedirectLink godoc
// @Summary Redirect to original URL
// @Description Redirects to the original URL and tracks analytics
// @Tags redirect
// @Param code path string true "Short code"
// @Success 302
// @Failure 404 {object} handler.ErrorEnvelope
// @Failure 410 {object} handler.ErrorEnvelope
// @Router /{code} [get]
func (h *LinkHandler) RedirectLink(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		RespondError(c, http.StatusBadRequest, CodeBadRequest, "Short code is required", nil)
		return
	}

	trackingData := h.extractTrackingData(c)

	originalURL, err := h.linkUseCase.RedirectLink(shortCode, trackingData)
	if err != nil {
		switch err {
		case domain.ErrLinkNotFound:
			RespondError(c, http.StatusNotFound, CodeNotFound, "Short link not found", nil)
		case domain.ErrLinkExpired:
			RespondError(c, http.StatusGone, CodeGone, "Link has expired", nil)
		case domain.ErrLinkInactive:
			RespondError(c, http.StatusForbidden, CodeForbidden, "Link is inactive", nil)
		case domain.ErrRateLimitExceeded:
			RespondError(c, http.StatusTooManyRequests, CodeTooManyRequests, "Rate limit exceeded", nil)
		default:
			if errors.Is(err, context.DeadlineExceeded) {
				RespondError(c, http.StatusGatewayTimeout, CodeGatewayTimeout, "Upstream timeout", nil)
			} else {
				RespondError(c, http.StatusInternalServerError, CodeInternalError, "An error occurred", nil)
			}
		}
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

// GenerateQRCode godoc
// @Summary Generate QR code
// @Description Generates a QR code for the short link
// @Tags qr
// @Param code path string true "Short code"
// @Produce image/png
// @Success 200
// @Failure 404 {object} handler.ErrorEnvelope
// @Failure 410 {object} handler.ErrorEnvelope
// @Router /qr/{code} [get]
func (h *LinkHandler) GenerateQRCode(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		RespondError(c, http.StatusBadRequest, CodeBadRequest, "Short code is required", nil)
		return
	}

	qrData, err := h.linkUseCase.GenerateQRCode(shortCode)
	if err != nil {
		switch err {
		case domain.ErrLinkNotFound:
			RespondError(c, http.StatusNotFound, CodeNotFound, "Short link not found", nil)
		case domain.ErrLinkExpired:
			RespondError(c, http.StatusGone, CodeGone, "Link has expired", nil)
		case domain.ErrLinkInactive:
			RespondError(c, http.StatusForbidden, CodeForbidden, "Link is inactive", nil)
		case domain.ErrRateLimitExceeded:
			RespondError(c, http.StatusTooManyRequests, CodeTooManyRequests, "Rate limit exceeded", nil)
		default:
			if errors.Is(err, context.DeadlineExceeded) {
				RespondError(c, http.StatusGatewayTimeout, CodeGatewayTimeout, "Upstream timeout", nil)
			} else {
				RespondError(c, http.StatusInternalServerError, CodeInternalError, "An error occurred", nil)
			}
		}
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=604800") // Cache for 7 days
	c.Data(http.StatusOK, "image/png", qrData)
}

// extractTrackingData extracts tracking information from the request
func (h *LinkHandler) extractTrackingData(c *gin.Context) *domain.TrackingData {
	clientIP := c.ClientIP()
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		clientIP = realIP
	}

	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")

	if referrer == "" {
		referrer = "Direct Access"
	}

	return &domain.TrackingData{
		IPAddress: clientIP,
		UserAgent: userAgent,
		Referrer:  referrer,
	}
}
