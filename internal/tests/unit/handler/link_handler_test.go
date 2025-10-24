package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/internal/handler"
	"service-short-link/internal/tests/mocks"
	"service-short-link/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLinkHandler_CreateLink_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Build real usecase with mocks
	linkRepo := &mocks.LinkRepositoryMock{}
	linkCache := &mocks.LinkCacheMock{}
	analyticsRepo := &mocks.AnalyticsRepositoryMock{}
	codeGen := &mocks.ShortCodeGeneratorMock{IsValidFunc: func(s string) bool { return true }}
	uniq := &mocks.UniqueCodeServiceMock{GenerateUniqueCodeFunc: func(generator domain.ShortCodeGenerator, repo domain.LinkRepository, length int) (string, error) {
		return "abcde", nil
	}}
	qrGen := &mocks.QRCodeGeneratorMock{GenerateFunc: func(data string, size int) ([]byte, error) { return []byte{1}, nil }}
	urlVal := &mocks.URLValidatorMock{NormalizeFunc: func(u string) (string, error) { return u, nil }, IsSafeURLFunc: func(u string) bool { return true }, IsValidFunc: func(u string) bool { return true }}
	ua := &mocks.UserAgentParserMock{}
	cfg := &mocks.ConfigServiceMock{GetIntFunc: func(key string) int {
		if key == "cache.ttl_links" {
			return 60
		}
		return 0
	}, GetStringFunc: func(key string) string {
		if key == "shortlink.base_url" {
			return "https://sho.rt"
		}
		return ""
	}}

	uc := usecase.NewLinkUseCase(linkRepo, linkCache, analyticsRepo, codeGen, uniq, qrGen, urlVal, ua, cfg)
	lh := handler.NewLinkHandler(uc)
	r.POST("/api/v1/links", lh.CreateLink)

	body, _ := json.Marshal(map[string]interface{}{"original_url": "https://example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body)))

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify response format
	var resp handler.SuccessEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)

	// Verify response data structure
	dataBytes, _ := json.Marshal(resp.Data)
	var data map[string]interface{}
	json.Unmarshal(dataBytes, &data)
	assert.Contains(t, data, "short_url")
	assert.Contains(t, data, "qr_url")
	assert.Equal(t, "https://sho.rt/abcde", data["short_url"])
	assert.Equal(t, "https://sho.rt/qr/abcde", data["qr_url"])
}

func TestLinkHandler_CreateLink_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	uc := usecase.NewLinkUseCase(&mocks.LinkRepositoryMock{}, &mocks.LinkCacheMock{}, &mocks.AnalyticsRepositoryMock{}, &mocks.ShortCodeGeneratorMock{}, &mocks.UniqueCodeServiceMock{}, &mocks.QRCodeGeneratorMock{}, &mocks.URLValidatorMock{}, &mocks.UserAgentParserMock{}, &mocks.ConfigServiceMock{})
	lh := handler.NewLinkHandler(uc)
	r.POST("/api/v1/links", lh.CreateLink)

	// invalid json
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString("{")))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error response format
	var resp handler.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, handler.CodeBadRequest, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Invalid request body")
}

func TestLinkHandler_Redirect_And_QR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	linkRepo := &mocks.LinkRepositoryMock{GetForRedirectFunc: func(shortCode string) (*domain.LinkForRedirect, error) {
		now := time.Now().Add(1 * time.Hour)
		return &domain.LinkForRedirect{ID: 1, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &now}, nil
	}}
	linkCache := &mocks.LinkCacheMock{}
	analyticsRepo := &mocks.AnalyticsRepositoryMock{}
	codeGen := &mocks.ShortCodeGeneratorMock{}
	uniq := &mocks.UniqueCodeServiceMock{}
	qrGen := &mocks.QRCodeGeneratorMock{GenerateFunc: func(data string, size int) ([]byte, error) { return []byte{1, 2, 3}, nil }}
	urlVal := &mocks.URLValidatorMock{IsValidFunc: func(u string) bool { return true }, NormalizeFunc: func(u string) (string, error) { return u, nil }, IsSafeURLFunc: func(u string) bool { return true }}
	ua := &mocks.UserAgentParserMock{}
	cfg := &mocks.ConfigServiceMock{GetIntFunc: func(key string) int {
		if key == "cache.ttl_links" || key == "cache.ttl_qr" {
			return 60
		}
		return 0
	}, GetStringFunc: func(key string) string {
		if key == "shortlink.base_url" {
			return "https://sho.rt"
		}
		return ""
	}}
	uc := usecase.NewLinkUseCase(linkRepo, linkCache, analyticsRepo, codeGen, uniq, qrGen, urlVal, ua, cfg)
	lh := handler.NewLinkHandler(uc)

	r.GET("/:code", lh.RedirectLink)
	r.GET("/qr/:code", lh.GenerateQRCode)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/abcde", nil))
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/qr/abcde", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=604800", w.Header().Get("Cache-Control"))
	assert.Equal(t, []byte{1, 2, 3}, w.Body.Bytes())
}

// Test error cases
func TestLinkHandler_CreateLink_InvalidURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock usecase to return ErrInvalidURL
	uc := &mocks.LinkUseCaseMock{
		CreateLinkFunc: func(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error) {
			return nil, domain.ErrInvalidURL
		},
	}
	lh := handler.NewLinkHandler(uc)
	r.POST("/api/v1/links", lh.CreateLink)

	body, _ := json.Marshal(map[string]interface{}{"original_url": "https://example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp handler.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, handler.CodeBadRequest, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Invalid URL format")
}

func TestLinkHandler_CreateLink_ShortCodeExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock usecase to return ErrShortCodeExists
	uc := &mocks.LinkUseCaseMock{
		CreateLinkFunc: func(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error) {
			return nil, domain.ErrShortCodeExists
		},
	}
	lh := handler.NewLinkHandler(uc)
	r.POST("/api/v1/links", lh.CreateLink)

	body, _ := json.Marshal(map[string]interface{}{
		"original_url": "https://example.com",
		"short_code":   "existing",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp handler.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, handler.CodeConflict, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Short code already exists")
}

func TestLinkHandler_Redirect_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock usecase to return ErrLinkNotFound
	uc := &mocks.LinkUseCaseMock{
		RedirectLinkFunc: func(shortCode string, trackingData *domain.TrackingData) (string, error) {
			return "", domain.ErrLinkNotFound
		},
	}
	lh := handler.NewLinkHandler(uc)
	r.GET("/:code", lh.RedirectLink)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/nonexistent", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp handler.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, handler.CodeNotFound, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Short link not found")
}

func TestLinkHandler_Redirect_Expired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock usecase to return ErrLinkExpired
	uc := &mocks.LinkUseCaseMock{
		RedirectLinkFunc: func(shortCode string, trackingData *domain.TrackingData) (string, error) {
			return "", domain.ErrLinkExpired
		},
	}
	lh := handler.NewLinkHandler(uc)
	r.GET("/:code", lh.RedirectLink)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/expired", nil))

	assert.Equal(t, http.StatusGone, w.Code)

	var resp handler.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, handler.CodeGone, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Link has expired")
}

// import time for expiration setup
