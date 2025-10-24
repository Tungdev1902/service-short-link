package domain_test

import (
	"testing"
	"service-short-link/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestErrors_Basic(t *testing.T) {
	imported := func() {}
	_ = imported

	assert.NotNil(t, domain.ErrLinkNotFound)
	assert.EqualError(t, domain.ErrLinkNotFound, "link not found")
	assert.NotNil(t, domain.ErrLinkExpired)
	assert.EqualError(t, domain.ErrLinkExpired, "link has expired")
	assert.NotNil(t, domain.ErrShortCodeExists)
	assert.EqualError(t, domain.ErrShortCodeExists, "short code already exists")
	assert.NotNil(t, domain.ErrAnalyticsNotFound)
	assert.EqualError(t, domain.ErrAnalyticsNotFound, "analytics data not found")
	assert.NotNil(t, domain.ErrCacheNotFound)
	assert.EqualError(t, domain.ErrCacheNotFound, "cache entry not found")
	assert.NotNil(t, domain.ErrInternalError)
	assert.EqualError(t, domain.ErrInternalError, "internal server error")
}
