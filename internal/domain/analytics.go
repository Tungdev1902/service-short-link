package domain

import (
	"time"
)

// Analytics represents tracking data for a short link access
type Analytics struct {
	ID        uint64    `json:"id" db:"id"`
	LinkID    uint64    `json:"link_id" db:"link_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent *string   `json:"user_agent,omitempty" db:"user_agent"`
	Referrer  *string   `json:"referrer,omitempty" db:"referrer"`
	Device    *string   `json:"device,omitempty" db:"device"`
	OS        *string   `json:"os,omitempty" db:"os"`
	Browser   *string   `json:"browser,omitempty" db:"browser"`
	ClickedAt time.Time `json:"clicked_at" db:"clicked_at"`
}

// TrackingData represents the data extracted from a request for analytics
type TrackingData struct {
	IPAddress string
	UserAgent string
	Referrer  string
	Device    string
	OS        string
	Browser   string
}

// AnalyticsRepository defines the interface for analytics data operations
type AnalyticsRepository interface {
	Create(analytics *Analytics) error
}

// UserAgentParser defines the interface for parsing user agent strings
type UserAgentParser interface {
	Parse(userAgent string) (device, os, browser string)
}