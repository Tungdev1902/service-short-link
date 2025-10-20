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
	// OriginalURL is required and must be a valid URL (max 2048 characters)
	OriginalURL string `json:"original_url" binding:"required,url" example:"https://example.com/very/long/path" validate:"required,url,max=2048"`

	// ShortCode is optional. If provided, must be 5-10 characters long, alphanumeric only, and not a reserved word
	ShortCode *string `json:"short_code,omitempty" example:"abc123" validate:"omitempty,min=5,max=10,alphanum"`

	// Title is optional (max 255 characters)
	Title *string `json:"title,omitempty" example:"Example Website" validate:"omitempty,max=255"`

	// Description is optional (max 1000 characters)
	Description *string `json:"description,omitempty" example:"This is an example website" validate:"omitempty,max=1000"`
}

// CreateLinkResponse represents the response after creating a short link
type CreateLinkResponse struct {
	ID          uint64     `json:"id" example:"1"`
	ShortCode   string     `json:"short_code" example:"aB3x9K2"`
	ShortURL    string     `json:"short_url" example:"http://short.vieclam24h.vn/aB3x9K2"`
	QRCodeURL   string     `json:"qr_url" example:"http://short.vieclam24h.vn/qr/aB3x9K2"`
	OriginalURL string     `json:"original_url" example:"https://example.com/very/long/path"`
	Title       *string    `json:"title,omitempty" example:"Example Website"`
	Description *string    `json:"description,omitempty" example:"This is an example website"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-15T10:30:45Z"`
}

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	Create(link *Link) error
	GetByShortCode(shortCode string) (*Link, error)
	ExistsByShortCode(shortCode string) (bool, error)
	IncrementClickCount(linkID uint64) error
	UpdateLastAccessed(linkID uint64) error
}

// LinkCache defines the interface for link caching operations
type LinkCache interface {
	Set(key string, link *Link, ttl time.Duration) error
	Get(key string) (*Link, error)
	SetQRCode(shortCode string, qrData []byte, ttl time.Duration) error
	GetQRCode(shortCode string) ([]byte, error)
}
