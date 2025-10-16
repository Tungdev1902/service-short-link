package domain

import "errors"

// Domain errors
var (
	// Link errors
	ErrLinkNotFound      = errors.New("link not found")
	ErrLinkExpired       = errors.New("link has expired")
	ErrLinkInactive      = errors.New("link is inactive")
	ErrShortCodeExists   = errors.New("short code already exists")
	ErrInvalidURL        = errors.New("invalid URL format")
	ErrInvalidShortCode  = errors.New("invalid short code format")
	ErrShortCodeTooLong  = errors.New("short code is too long")
	ErrShortCodeTooShort = errors.New("short code is too short")

	// Analytics errors
	ErrAnalyticsNotFound = errors.New("analytics data not found")
	ErrInvalidIPAddress  = errors.New("invalid IP address")

	// Cache errors
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheExpired  = errors.New("cache entry has expired")

	// Authentication errors
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token has expired")
	ErrInvalidAPIKey    = errors.New("invalid API key")
	ErrMissingAuthToken = errors.New("missing authorization token")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// QR Code errors
	ErrQRCodeGeneration = errors.New("failed to generate QR code")
	ErrQRCodeNotFound   = errors.New("QR code not found")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")

	// Configuration errors
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrMissingConfiguration = errors.New("missing required configuration")

	// Generic internal error for security (doesn't leak details)
	ErrInternalError = errors.New("internal server error")
)