package handler

import (
    "net/http"

    "service-short-link/internal/domain"
    "service-short-link/internal/usecase"
    "service-short-link/internal/handler/transform"

    "github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkUseCase *usecase.LinkUseCase
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(linkUseCase *usecase.LinkUseCase) *LinkHandler {
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
// @Success 201 {object} handler.SuccessEnvelope{data=transform.LinkCreatedDTO} "Link created successfully"
// @Failure 400 {object} handler.ErrorEnvelope "Bad request - invalid input or short code already exists"
// @Failure 401 {object} handler.ErrorEnvelope "Unauthorized - valid JWT token or API key required"
// @Failure 500 {object} handler.ErrorEnvelope "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /api/v1/links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req domain.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if len(req.OriginalURL) > 2048 {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "URL too long (max 2048 characters)", nil)
		return
	}
	if req.Title != nil && len(*req.Title) > 255 {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Title too long (max 255 characters)", nil)
		return
	}
	if req.Description != nil && len(*req.Description) > 1000 {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Description too long (max 1000 characters)", nil)
		return
	}
	
	if req.ShortCode != nil && *req.ShortCode != "" {
		if len(*req.ShortCode) < 5 {
			RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Short code too short (min 5 characters)", nil)
			return
		}
		if len(*req.ShortCode) > 10 {
			RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Short code too long (max 10 characters)", nil)
			return
		}
		for _, char := range *req.ShortCode {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Short code must contain only alphanumeric characters (a-z, A-Z, 0-9)", nil)
				return
			}
		}
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
            RespondError(c, http.StatusBadRequest, "INVALID_URL", "Invalid URL format", nil)
		case domain.ErrShortCodeExists:
            RespondError(c, http.StatusBadRequest, "SHORT_CODE_EXISTS", "Short code already exists", nil)
		case domain.ErrInvalidShortCode:
            RespondError(c, http.StatusBadRequest, "INVALID_SHORT_CODE", "Invalid short code format - must be 5-10 alphanumeric characters and not a reserved word", nil)
		default:
            RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred", nil)
		}
		return
	}

    dto := transform.ToLinkCreatedDTO(response)
    RespondCreated(c, dto)
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
        RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Short code is required", nil)
		return
	}

	trackingData := h.extractTrackingData(c)

	originalURL, err := h.linkUseCase.RedirectLink(shortCode, trackingData)
	if err != nil {
		switch err {
		case domain.ErrLinkNotFound:
            RespondError(c, http.StatusNotFound, "NOT_FOUND", "Short link not found", nil)
		case domain.ErrLinkExpired:
            RespondError(c, http.StatusGone, "LINK_EXPIRED", "Link has expired", nil)
		case domain.ErrLinkInactive:
            RespondError(c, http.StatusForbidden, "LINK_INACTIVE", "Link is inactive", nil)
		default:
            RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred", nil)
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
        RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "Short code is required", nil)
		return
	}

	qrData, err := h.linkUseCase.GenerateQRCode(shortCode)
	if err != nil {
		switch err {
		case domain.ErrLinkNotFound:
            RespondError(c, http.StatusNotFound, "NOT_FOUND", "Short link not found", nil)
		case domain.ErrLinkExpired:
            RespondError(c, http.StatusGone, "LINK_EXPIRED", "Link has expired", nil)
		case domain.ErrLinkInactive:
            RespondError(c, http.StatusForbidden, "LINK_INACTIVE", "Link is inactive", nil)
		default:
            RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred", nil)
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

