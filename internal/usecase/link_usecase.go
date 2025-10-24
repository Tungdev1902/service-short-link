package usecase

import (
	"fmt"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"
)

type LinkUseCase struct {
	linkRepo          domain.LinkRepository
	linkCache         domain.LinkCache
	analyticsRepo     domain.AnalyticsRepository
	codeGenerator     domain.ShortCodeGenerator
	uniqueCodeService domain.UniqueCodeService
	qrGenerator       domain.QRCodeGenerator
	urlValidator      domain.URLValidator
	userAgentParser   domain.UserAgentParser
	configService     domain.ConfigService
}

// NewLinkUseCase creates a new link use case
func NewLinkUseCase(
	linkRepo domain.LinkRepository,
	linkCache domain.LinkCache,
	analyticsRepo domain.AnalyticsRepository,
	codeGenerator domain.ShortCodeGenerator,
	uniqueCodeService domain.UniqueCodeService,
	qrGenerator domain.QRCodeGenerator,
	urlValidator domain.URLValidator,
	userAgentParser domain.UserAgentParser,
	configService domain.ConfigService,
) *LinkUseCase {
	return &LinkUseCase{
		linkRepo:          linkRepo,
		linkCache:         linkCache,
		analyticsRepo:     analyticsRepo,
		codeGenerator:     codeGenerator,
		uniqueCodeService: uniqueCodeService,
		qrGenerator:       qrGenerator,
		urlValidator:      urlValidator,
		userAgentParser:   userAgentParser,
		configService:     configService,
	}
}

// CreateLink creates a new short link with race condition handling
func (uc *LinkUseCase) CreateLink(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error) {
	normalizedURL, err := uc.urlValidator.Normalize(req.OriginalURL)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "CreateLink: failed to normalize URL", "original_url="+req.OriginalURL, "error_type=url_validation_failed")
		return nil, domain.ErrInvalidURL
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
		shortCode, err = uc.uniqueCodeService.GenerateUniqueCode(
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
	seconds := uc.configService.GetInt("shortlink.default_expiry_seconds")

	if seconds > 0 {
		expiry := time.Now().Add(time.Duration(seconds) * time.Second)
		expiresAt = &expiry
		logger.Info("CreateLink: setting expiry", map[string]interface{}{
			"expires_at": expiry.Format(time.RFC3339),
		})
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
	err = uc.linkCache.Set(shortCode, link.ToCache(), ttl)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "CreateLink: failed to cache link", "short_code="+shortCode, "error_type=cache_set_failed")
	}

	baseURL := uc.configService.GetString("shortlink.base_url")
	response := &domain.CreateLinkResponse{
		ShortURL:  fmt.Sprintf("%s/%s", baseURL, link.ShortCode),
		QRCodeURL: fmt.Sprintf("%s/qr/%s", baseURL, link.ShortCode),
	}

	return response, nil
}

// RedirectLink handles link redirection with analytics tracking
func (uc *LinkUseCase) RedirectLink(shortCode string, trackingData *domain.TrackingData) (string, error) {
	if cachedLink, err := uc.linkCache.Get(shortCode); err == nil && cachedLink != nil {
		return uc.handleCacheHit(cachedLink, trackingData)
	}

	return uc.handleCacheMiss(shortCode, trackingData)
}

// handleCacheHit processes redirect when link is found in cache
func (uc *LinkUseCase) handleCacheHit(cachedLink *domain.CachedLink, trackingData *domain.TrackingData) (string, error) {
	if err := uc.checkLinkAccessibility(cachedLink.IsActive, cachedLink.ExpiresAt); err != nil {
		return "", err
	}

	uc.trackLinkUsage(cachedLink.ID, trackingData)

	return cachedLink.OriginalURL, nil
}

// handleCacheMiss processes redirect when link is not in cache
func (uc *LinkUseCase) handleCacheMiss(shortCode string, trackingData *domain.TrackingData) (string, error) {
	link, err := uc.linkRepo.GetForRedirect(shortCode)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "RedirectLink: failed to get link from database",
			"short_code", shortCode, "error_type", "database_query_failed")
		return "", err
	}

	if !link.IsAccessible() {
		if link.IsExpired() {
			return "", domain.ErrLinkExpired
		}
		return "", domain.ErrLinkInactive
	}

	uc.updateLinkCacheFromRedirect(shortCode, link)

	uc.trackLinkUsage(link.ID, trackingData)

	return link.OriginalURL, nil
}

// checkLinkAccessibility validates if link can be accessed
func (uc *LinkUseCase) checkLinkAccessibility(isActive bool, expiresAt *time.Time) error {
	if !isActive {
		return domain.ErrLinkInactive
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		return domain.ErrLinkExpired
	}

	return nil
}

// updateLinkCacheFromRedirect stores redirect link in cache with configured TTL
func (uc *LinkUseCase) updateLinkCacheFromRedirect(shortCode string, link *domain.LinkForRedirect) {
	ttl := time.Duration(uc.configService.GetInt("cache.ttl_links")) * time.Second
	if err := uc.linkCache.Set(shortCode, link.ToCache(), ttl); err != nil {
		logger.ErrorWithCockroachSimple(err, "RedirectLink: failed to update cache",
			"short_code", shortCode, "error_type", "cache_set_failed")
	}
}

// trackLinkUsage handles all analytics tracking asynchronously
func (uc *LinkUseCase) trackLinkUsage(linkID uint64, trackingData *domain.TrackingData) {
	go func() {
		linkIDStr := fmt.Sprintf("%d", linkID)

		if err := uc.linkRepo.IncrementClickCount(linkID); err != nil {
			logger.ErrorWithCockroachSimple(err, "Failed to increment click count",
				"link_id", linkIDStr)
		}

		if err := uc.linkRepo.UpdateLastAccessed(linkID); err != nil {
			logger.ErrorWithCockroachSimple(err, "Failed to update last accessed",
				"link_id", linkIDStr)
		}

		uc.trackAnalytics(linkID, trackingData)
	}()
}

// GenerateQRCode generates QR code for a short link
func (uc *LinkUseCase) GenerateQRCode(shortCode string) ([]byte, error) {
	qrData, err := uc.linkCache.GetQRCode(shortCode)
	if err == nil && qrData != nil {
		return qrData, nil
	}

	link, err := uc.linkRepo.GetForRedirect(shortCode)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "GenerateQRCode: failed to get link from database",
			"short_code", shortCode, "error_type", "database_query_failed")
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
		logger.ErrorWithCockroachSimple(err, "GenerateQRCode: failed to generate QR",
			"short_code", shortCode, "short_url", shortURL)
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
