package domain_test

import (
	"service-short-link/internal/domain"
	"service-short-link/internal/tests/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLink_Basic(t *testing.T) {
	// Test IsExpired: ExpiresAt nil
	link := &domain.Link{IsActive: true}
	assert.False(t, link.IsExpired())

	// Test IsExpired: ExpiresAt in future
	future := time.Now().Add(1 * time.Hour)
	link.ExpiresAt = &future
	assert.False(t, link.IsExpired())

	// Test IsExpired: ExpiresAt in past
	past := time.Now().Add(-1 * time.Hour)
	link.ExpiresAt = &past
	assert.True(t, link.IsExpired())

	// Test IsAccessible: active & not expired
	link.IsActive = true
	link.ExpiresAt = &future
	assert.True(t, link.IsAccessible())

	// Test IsAccessible: inactive
	link.IsActive = false
	assert.False(t, link.IsAccessible())

	// Test IsAccessible: expired
	link.IsActive = true
	link.ExpiresAt = &past
	assert.False(t, link.IsAccessible())

	// Test ToCache
	link.ID = 123
	link.OriginalURL = "https://example.com"
	cache := link.ToCache()
	assert.Equal(t, link.ID, cache.ID)
	assert.Equal(t, link.OriginalURL, cache.OriginalURL)
	assert.Equal(t, link.IsActive, cache.IsActive)
	assert.Equal(t, link.ExpiresAt, cache.ExpiresAt)
}

func TestLinkForRedirect_Basic(t *testing.T) {
	link := &domain.LinkForRedirect{IsActive: true}
	assert.False(t, link.IsExpired())

	future := time.Now().Add(1 * time.Hour)
	link.ExpiresAt = &future
	assert.False(t, link.IsExpired())

	past := time.Now().Add(-1 * time.Hour)
	link.ExpiresAt = &past
	assert.True(t, link.IsExpired())

	link.IsActive = true
	link.ExpiresAt = &future
	assert.True(t, link.IsAccessible())
	link.IsActive = false
	assert.False(t, link.IsAccessible())
	link.IsActive = true
	link.ExpiresAt = &past
	assert.False(t, link.IsAccessible())

	cache := link.ToCache()
	assert.Equal(t, link.ID, cache.ID)
	assert.Equal(t, link.OriginalURL, cache.OriginalURL)
	assert.Equal(t, link.IsActive, cache.IsActive)
	assert.Equal(t, link.ExpiresAt, cache.ExpiresAt)
}

func TestCachedLink_Basic(t *testing.T) {
	cl := &domain.CachedLink{IsActive: true}
	assert.False(t, cl.IsExpired())
	future := time.Now().Add(1 * time.Hour)
	cl.ExpiresAt = &future
	assert.False(t, cl.IsExpired())
	past := time.Now().Add(-1 * time.Hour)
	cl.ExpiresAt = &past
	assert.True(t, cl.IsExpired())

	cl.IsActive = true
	cl.ExpiresAt = &future
	assert.True(t, cl.IsAccessible())
	cl.IsActive = false
	assert.False(t, cl.IsAccessible())
	cl.IsActive = true
	cl.ExpiresAt = &past
	assert.False(t, cl.IsAccessible())
}

func TestLinkRepositoryMock(t *testing.T) {
	mockRepo := &mocks.LinkRepositoryMock{
		CreateFunc: func(link *domain.Link) error {
			if link.ShortCode == "abc" {
				return nil
			}
			return assert.AnError
		},
		GetByShortCodeFunc: func(shortCode string) (*domain.Link, error) {
			if shortCode == "abc" {
				return &domain.Link{ShortCode: "abc"}, nil
			}
			return nil, assert.AnError
		},
	}

	err := mockRepo.Create(&domain.Link{ShortCode: "abc"})
	assert.NoError(t, err)

	link, err := mockRepo.GetByShortCode("abc")
	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, "abc", link.ShortCode)
}
