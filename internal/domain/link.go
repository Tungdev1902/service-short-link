package domain

import (
	"time"
)

// Link represents a shortened URL
type Link struct {
	ID             uint64     `json:"id" db:"id"`
	ShortCode      string     `json:"short_code" db:"short_code"`
	OriginalURL    string     `json:"original_url" db:"original_url"`
	Title          *string    `json:"title,omitempty" db:"title"`
	Description    *string    `json:"description,omitempty" db:"description"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Platform       *string    `json:"platform,omitempty" db:"platform"`
	Role           *string    `json:"role,omitempty" db:"role"`
	ChannelCode    *string    `json:"channel_code,omitempty" db:"channel_code"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	ClickCount     uint64     `json:"click_count" db:"click_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty" db:"last_accessed_at"`
}

// IsExpired checks if the link has expired
func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// IsAccessible checks if the link can be accessed
func (l *Link) IsAccessible() bool {
	return l.IsActive && !l.IsExpired()
}

// CreateLinkRequest represents the request to create a new short link
type CreateLinkRequest struct {
	OriginalURL string  `json:"original_url" binding:"required,url" example:"https://example.com/very/long/path" validate:"required,url,max=2048"`
	ShortCode   *string `json:"short_code,omitempty" example:"abc123" validate:"omitempty,min=5,max=10,alphanum"`
	Title       *string `json:"title,omitempty" example:"Example Website" validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty" example:"This is an example website" validate:"omitempty,max=1000"`
}

// CreateLinkResponse represents the response after creating a short link
type CreateLinkResponse struct {
	ShortURL  string `json:"short_url" example:"http://short.vieclam24h.vn/aB3x9K2"`
	QRCodeURL string `json:"qr_url" example:"http://short.vieclam24h.vn/qr/aB3x9K2"`
}

// LinkForRedirect contains minimal data needed for redirection
type LinkForRedirect struct {
	ID          uint64     `json:"id" db:"id"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// IsExpired checks if the link has expired
func (l *LinkForRedirect) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// IsAccessible checks if the link can be accessed
func (l *LinkForRedirect) IsAccessible() bool {
	return l.IsActive && !l.IsExpired()
}

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	Create(link *Link) error
	GetByShortCode(shortCode string) (*Link, error)
	GetForRedirect(shortCode string) (*LinkForRedirect, error)
	ExistsByShortCode(shortCode string) (bool, error)
	IncrementClickCount(linkID uint64) error
	UpdateLastAccessed(linkID uint64) error
}

// CachedLink represents link data for caching (same as LinkForRedirect)
type CachedLink struct {
	ID          uint64     `json:"id"`
	OriginalURL string     `json:"original_url"`
	IsActive    bool       `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ToCache converts LinkForRedirect to CachedLink
func (l *LinkForRedirect) ToCache() *CachedLink {
	return &CachedLink{
		ID:          l.ID,
		OriginalURL: l.OriginalURL,
		IsActive:    l.IsActive,
		ExpiresAt:   l.ExpiresAt,
	}
}

// ToCacheFromLink converts full Link to CachedLink (for create operation)
func (l *Link) ToCache() *CachedLink {
	return &CachedLink{
		ID:          l.ID,
		OriginalURL: l.OriginalURL,
		IsActive:    l.IsActive,
		ExpiresAt:   l.ExpiresAt,
	}
}

// IsExpired checks if the cached link has expired
func (cl *CachedLink) IsExpired() bool {
	if cl.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*cl.ExpiresAt)
}

// IsAccessible checks if the cached link can be accessed
func (cl *CachedLink) IsAccessible() bool {
	return cl.IsActive && !cl.IsExpired()
}

// LinkCache defines the interface for link caching operations
type LinkCache interface {
	Set(key string, link *CachedLink, ttl time.Duration) error
	Get(key string) (*CachedLink, error)
	SetQRCode(shortCode string, qrData []byte, ttl time.Duration) error
	GetQRCode(shortCode string) ([]byte, error)
}
