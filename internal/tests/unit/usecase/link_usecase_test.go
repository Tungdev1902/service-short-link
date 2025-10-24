package usecase_test

import (
	"testing"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/internal/tests/mocks"
	"service-short-link/internal/usecase"

	"github.com/stretchr/testify/assert"
)

func buildDefaultUseCase() (*usecase.LinkUseCase, *mocks.LinkRepositoryMock, *mocks.LinkCacheMock, *mocks.AnalyticsRepositoryMock, *mocks.ShortCodeGeneratorMock, *mocks.UniqueCodeServiceMock, *mocks.QRCodeGeneratorMock, *mocks.URLValidatorMock, *mocks.UserAgentParserMock, *mocks.ConfigServiceMock) {
	linkRepo := &mocks.LinkRepositoryMock{}
	linkCache := &mocks.LinkCacheMock{}
	analyticsRepo := &mocks.AnalyticsRepositoryMock{}
	codeGen := &mocks.ShortCodeGeneratorMock{IsValidFunc: func(s string) bool { return true }}
	uniq := &mocks.UniqueCodeServiceMock{GenerateUniqueCodeFunc: func(generator domain.ShortCodeGenerator, repo domain.LinkRepository, length int) (string, error) {
		return "abcde", nil
	}}
	qrGen := &mocks.QRCodeGeneratorMock{GenerateFunc: func(data string, size int) ([]byte, error) { return []byte{9}, nil }}
	urlVal := &mocks.URLValidatorMock{NormalizeFunc: func(u string) (string, error) { return u, nil }, IsSafeURLFunc: func(u string) bool { return true }, IsValidFunc: func(u string) bool { return true }}
	ua := &mocks.UserAgentParserMock{ParseFunc: func(userAgent string) (device, os, browser string) { return "m", "o", "b" }}
	cfg := &mocks.ConfigServiceMock{GetIntFunc: func(key string) int {
		if key == "cache.ttl_links" || key == "cache.ttl_qr" {
			return 60
		}
		if key == "shortlink.shortcode_length" {
			return 5
		}
		if key == "shortlink.default_expiry_seconds" {
			return 3600
		}
		return 0
	}, GetStringFunc: func(key string) string {
		if key == "shortlink.base_url" {
			return "https://sho.rt"
		}
		return ""
	}}
	return usecase.NewLinkUseCase(linkRepo, linkCache, analyticsRepo, codeGen, uniq, qrGen, urlVal, ua, cfg), linkRepo, linkCache, analyticsRepo, codeGen, uniq, qrGen, urlVal, ua, cfg
}

func TestCreateLink_Success(t *testing.T) {
	uc, linkRepo, linkCache, _, _, _, _, _, _, cfg := buildDefaultUseCase()

	// capture create and cache set
	created := false
	cached := false
	linkRepo.CreateFunc = func(link *domain.Link) error { created = true; link.ID = 1; return nil }
	linkCache.SetFunc = func(key string, link *domain.CachedLink, ttl time.Duration) error {
		cached = true
		assert.Equal(t, 60*time.Second, ttl)
		return nil
	}

	req := &domain.CreateLinkRequest{OriginalURL: "https://example.com"}
	resp, err := uc.CreateLink(req, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.ShortURL, "/")
	assert.True(t, created)
	assert.True(t, cached)

	// ensure expiry used when configured
	assert.NotZero(t, cfg.GetInt("shortlink.default_expiry_seconds"))
}

func TestCreateLink_InvalidShortCode(t *testing.T) {
	uc, _, _, _, codeGen, _, _, _, _, _ := buildDefaultUseCase()
	code := "bad"
	codeGen.IsValidFunc = func(s string) bool { return false }
	req := &domain.CreateLinkRequest{OriginalURL: "https://example.com", ShortCode: &code}
	resp, err := uc.CreateLink(req, nil, nil, nil)
	assert.Nil(t, resp)
	assert.Equal(t, domain.ErrInvalidShortCode, err)
}

func TestRedirectLink_CacheHit(t *testing.T) {
	uc, _, linkCache, _, _, _, _, _, _, _ := buildDefaultUseCase()
	now := time.Now().Add(1 * time.Hour)
	linkCache.GetFunc = func(key string) (*domain.CachedLink, error) {
		return &domain.CachedLink{ID: 1, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &now}, nil
	}

	url, err := uc.RedirectLink("abcde", &domain.TrackingData{UserAgent: "ua"})
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", url)
}

func TestRedirectLink_CacheMiss_DBExpired(t *testing.T) {
	uc, linkRepo, _, _, _, _, _, _, _, _ := buildDefaultUseCase()
	expired := time.Now().Add(-1 * time.Hour)
	linkRepo.GetForRedirectFunc = func(shortCode string) (*domain.LinkForRedirect, error) {
		return &domain.LinkForRedirect{ID: 1, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &expired}, nil
	}

	_, err := uc.RedirectLink("abcde", &domain.TrackingData{})
	assert.Equal(t, domain.ErrLinkExpired, err)
}

func TestGenerateQRCode_CacheHit(t *testing.T) {
	uc, _, linkCache, _, _, _, _, _, _, _ := buildDefaultUseCase()
	linkCache.GetQRCodeFunc = func(shortCode string) ([]byte, error) { return []byte{1, 2}, nil }
	data, err := uc.GenerateQRCode("abcde")
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2}, data)
}

func TestGenerateQRCode_DBAndCache(t *testing.T) {
	uc, linkRepo, linkCache, _, _, _, qrGen, _, _, _ := buildDefaultUseCase()
	now := time.Now().Add(1 * time.Hour)
	linkRepo.GetForRedirectFunc = func(shortCode string) (*domain.LinkForRedirect, error) {
		return &domain.LinkForRedirect{ID: 1, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &now}, nil
	}
	linkCache.SetQRCodeFunc = func(shortCode string, qrData []byte, ttl time.Duration) error {
		assert.Equal(t, 60*time.Second, ttl)
		return nil
	}
	qrGen.GenerateFunc = func(data string, size int) ([]byte, error) { return []byte{9, 9}, nil }

	data, err := uc.GenerateQRCode("abcde")
	assert.NoError(t, err)
	assert.Equal(t, []byte{9, 9}, data)
}

func TestRedirectLink_TracksAnalytics_OnCacheHit(t *testing.T) {
	uc, linkRepo, linkCache, analyticsRepo, _, _, _, _, _, _ := buildDefaultUseCase()

	// prepare cache hit
	future := time.Now().Add(1 * time.Hour)
	linkCache.GetFunc = func(key string) (*domain.CachedLink, error) {
		return &domain.CachedLink{ID: 42, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &future}, nil
	}

	// channels to wait for async tracking
	createdCh := make(chan *domain.Analytics, 1)
	clicksCh := make(chan uint64, 1)
	lastAccessCh := make(chan uint64, 1)

	analyticsRepo.CreateFunc = func(a *domain.Analytics) error {
		createdCh <- a
		return nil
	}
	linkRepo.IncrementClickCountFunc = func(linkID uint64) error {
		clicksCh <- linkID
		return nil
	}
	linkRepo.UpdateLastAccessedFunc = func(linkID uint64) error {
		lastAccessCh <- linkID
		return nil
	}

	// act
	url, err := uc.RedirectLink("abcde", &domain.TrackingData{UserAgent: "UA", IPAddress: "1.2.3.4", Referrer: "ref"})
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", url)

	// assert async tracking signals
	select {
	case a := <-createdCh:
		assert.Equal(t, uint64(42), a.LinkID)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for analytics create")
	}
	select {
	case id := <-clicksCh:
		assert.Equal(t, uint64(42), id)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for click increment")
	}
	select {
	case id := <-lastAccessCh:
		assert.Equal(t, uint64(42), id)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for last accessed update")
	}
}

func TestCreateLink_RepoExistsCheckError(t *testing.T) {
	uc, linkRepo, _, _, codeGen, _, _, _, _, _ := buildDefaultUseCase()
	code := "abcde"
	codeGen.IsValidFunc = func(s string) bool { return true }
	linkRepo.ExistsByShortCodeFunc = func(shortCode string) (bool, error) { return false, assert.AnError }
	req := &domain.CreateLinkRequest{OriginalURL: "https://example.com", ShortCode: &code}
	resp, err := uc.CreateLink(req, nil, nil, nil)
	assert.Nil(t, resp)
	assert.Equal(t, domain.ErrInternalError, err)
}

func TestCreateLink_RepoCreateError(t *testing.T) {
	uc, linkRepo, _, _, _, _, _, _, _, _ := buildDefaultUseCase()
	linkRepo.CreateFunc = func(link *domain.Link) error { return assert.AnError }
	resp, err := uc.CreateLink(&domain.CreateLinkRequest{OriginalURL: "https://example.com"}, nil, nil, nil)
	assert.Nil(t, resp)
	assert.Equal(t, domain.ErrInternalError, err)
}

func TestRedirectLink_Inactive(t *testing.T) {
	uc, linkRepo, _, _, _, _, _, _, _, _ := buildDefaultUseCase()
	now := time.Now().Add(1 * time.Hour)
	linkRepo.GetForRedirectFunc = func(shortCode string) (*domain.LinkForRedirect, error) {
		return &domain.LinkForRedirect{ID: 1, OriginalURL: "https://example.com", IsActive: false, ExpiresAt: &now}, nil
	}
	_, err := uc.RedirectLink("abcde", &domain.TrackingData{})
	assert.Equal(t, domain.ErrLinkInactive, err)
}

func TestGenerateQRCode_Inactive(t *testing.T) {
	uc, linkRepo, _, _, _, _, _, _, _, _ := buildDefaultUseCase()
	now := time.Now().Add(1 * time.Hour)
	linkRepo.GetForRedirectFunc = func(shortCode string) (*domain.LinkForRedirect, error) {
		return &domain.LinkForRedirect{ID: 1, OriginalURL: "https://example.com", IsActive: false, ExpiresAt: &now}, nil
	}
	_, err := uc.GenerateQRCode("abcde")
	assert.Equal(t, domain.ErrLinkInactive, err)
}
