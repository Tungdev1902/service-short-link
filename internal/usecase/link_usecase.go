package usecase

import (
	"fmt"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/internal/infrastructure/services"
	"service-short-link/pkg/logger"
)

type LinkUseCase struct {
	linkRepo        domain.LinkRepository
	linkCache       domain.LinkCache
	analyticsRepo   domain.AnalyticsRepository
	codeGenerator   domain.ShortCodeGenerator
	qrGenerator     domain.QRCodeGenerator
	urlValidator    domain.URLValidator
	userAgentParser domain.UserAgentParser
	configService   domain.ConfigService
}

// NewLinkUseCase creates a new link use case
func NewLinkUseCase(
	linkRepo domain.LinkRepository,
	linkCache domain.LinkCache,
	analyticsRepo domain.AnalyticsRepository,
	codeGenerator domain.ShortCodeGenerator,
	qrGenerator domain.QRCodeGenerator,
	urlValidator domain.URLValidator,
	userAgentParser domain.UserAgentParser,
	configService domain.ConfigService,
) *LinkUseCase {
	return &LinkUseCase{
		linkRepo:        linkRepo,
		linkCache:       linkCache,
		analyticsRepo:   analyticsRepo,
		codeGenerator:   codeGenerator,
		qrGenerator:     qrGenerator,
		urlValidator:    urlValidator,
		userAgentParser: userAgentParser,
		configService:   configService,
	}
}

// CreateLink creates a new short link with race condition handling
func (uc *LinkUseCase) CreateLink(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error) {
	normalizedURL, err := uc.urlValidator.Normalize(req.OriginalURL)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "CreateLink: failed to normalize URL", "original_url="+req.OriginalURL, "error_type=url_validation_failed")
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if !uc.urlValidator.IsSafeURL(normalizedURL) {
		return nil, domain.ErrInvalidURL
	}

	var shortCode string
	if req.ShortCode != nil && *req.ShortCode != "" {
		if !uc.codeGenerator.IsValid(*req.ShortCode) {
			return nil, domain.ErrInvalidShortCode
		}

		exists, err := uc.linkRepo.ExistsByShortCode(*req.ShortCode)
		if err != nil {
			logger.ErrorWithCockroachSimple(err, "CreateLink: failed to check short code existence", "error_type=database_query_failed")
			return nil, domain.ErrInternalError
		}
		if exists {
			return nil, domain.ErrShortCodeExists
		}

		shortCode = *req.ShortCode
	} else {
		shortCode, err = services.GenerateUniqueCode(
			uc.codeGenerator,
			uc.linkRepo,
			uc.configService.GetInt("shortlink.shortcode_length"),
		)
		if err != nil {
			logger.ErrorWithCockroachSimple(err, "CreateLink: failed to generate unique short code", "error_type=code_generation_failed")
			return nil, domain.ErrInternalError
		}
	}

	var expiresAt *time.Time
	// Use integer seconds; 0 or unset means no expiry
	if seconds := uc.configService.GetInt("shortlink.default_expiry_seconds"); seconds > 0 {
		expiry := time.Now().Add(time.Duration(seconds) * time.Second)
		expiresAt = &expiry
	}

	// Create link entity
	link := &domain.Link{
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		Title:       req.Title,
		Description: req.Description,
		IsActive:    true,
		ExpiresAt:   expiresAt,
		Platform:    platform,
		Role:        role,
		ChannelCode: channelCode,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ClickCount:  0,
	}

	err = uc.linkRepo.Create(link)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "CreateLink: failed to create link", "short_code="+shortCode, "error_type=database_create_failed")
		return nil, domain.ErrInternalError
	}

	ttl := time.Duration(uc.configService.GetInt("cache.ttl_links")) * time.Second
	err = uc.linkCache.Set(shortCode, link, ttl)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "CreateLink: failed to cache link", "short_code="+shortCode, "error_type=cache_set_failed")
	}

	baseURL := uc.configService.GetString("shortlink.base_url")
	response := &domain.CreateLinkResponse{
		ID:          link.ID,
		ShortCode:   link.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", baseURL, link.ShortCode),
		QRCodeURL:   fmt.Sprintf("%s/qr/%s", baseURL, link.ShortCode),
		OriginalURL: link.OriginalURL,
		Title:       link.Title,
		Description: link.Description,
		ExpiresAt:   link.ExpiresAt,
		CreatedAt:   link.CreatedAt,
	}

	return response, nil
}

// RedirectLink handles link redirection with analytics tracking
func (uc *LinkUseCase) RedirectLink(shortCode string, trackingData *domain.TrackingData) (string, error) {
	link, err := uc.linkCache.Get(shortCode)
	if err != nil || link == nil {
		link, err = uc.linkRepo.GetByShortCode(shortCode)
		if err != nil {
			logger.ErrorWithCockroachSimple(err, "RedirectLink: failed to get link from database", "short_code="+shortCode, "error_type=database_query_failed")
			return "", err
		}

		// Update cache for next time
		ttl := time.Duration(uc.configService.GetInt("cache.ttl_links")) * time.Second
		uc.linkCache.Set(shortCode, link, ttl)
	}

	if !link.IsAccessible() {
		if link.IsExpired() {
			return "", domain.ErrLinkExpired
		}
		return "", domain.ErrLinkInactive
	}

	go func() {
		uc.linkRepo.IncrementClickCount(link.ID)
		uc.linkRepo.UpdateLastAccessed(link.ID)
	}()

	go func() {
		uc.trackAnalytics(link.ID, trackingData)
	}()

	return link.OriginalURL, nil
}

// GenerateQRCode generates QR code for a short link
func (uc *LinkUseCase) GenerateQRCode(shortCode string) ([]byte, error) {
	qrData, err := uc.linkCache.GetQRCode(shortCode)
	if err == nil && qrData != nil {
		return qrData, nil
	}

	link, err := uc.linkRepo.GetByShortCode(shortCode)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "GenerateQRCode: failed to get link from database", "short_code="+shortCode, "error_type=database_query_failed")
		return nil, err
	}

	if !link.IsAccessible() {
		if link.IsExpired() {
			return nil, domain.ErrLinkExpired
		}
		return nil, domain.ErrLinkInactive
	}

	baseURL := uc.configService.GetString("shortlink.base_url")
	shortURL := fmt.Sprintf("%s/%s", baseURL, shortCode)

	qrData, err = uc.qrGenerator.Generate(shortURL, 256)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "GenerateQRCode: failed to generate QR", "short_code="+shortCode, "short_url="+shortURL)
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	ttl := time.Duration(uc.configService.GetInt("cache.ttl_qr")) * time.Second
	uc.linkCache.SetQRCode(shortCode, qrData, ttl)

	return qrData, nil
}

// trackAnalytics records analytics data for a link access
func (uc *LinkUseCase) trackAnalytics(linkID uint64, trackingData *domain.TrackingData) {
	if trackingData == nil {
		logger.Warn("trackingData is nil", map[string]interface{}{
			"link_id": linkID,
		})
		return
	}

	device, os, browser := uc.userAgentParser.Parse(trackingData.UserAgent)

	analytics := &domain.Analytics{
		LinkID:    linkID,
		IPAddress: truncateString(trackingData.IPAddress, 45),
		UserAgent: truncateStringPtr(&trackingData.UserAgent, 65535), // TEXT field
		Referrer:  truncateStringPtr(&trackingData.Referrer, 65535),  // TEXT field
		Device:    truncateStringPtr(&device, 50),
		OS:        truncateStringPtr(&os, 50),
		Browser:   truncateStringPtr(&browser, 50),
		ClickedAt: time.Now(),
	}

	err := uc.analyticsRepo.Create(analytics)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "trackAnalytics: failed to save analytics data", "link_id="+fmt.Sprintf("%d", linkID))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// truncateStringPtr truncates a string pointer to the specified length
func truncateStringPtr(s *string, maxLen int) *string {
	if s == nil {
		return nil
	}
	if len(*s) <= maxLen {
		return s
	}
	truncated := (*s)[:maxLen]
	return &truncated
}
