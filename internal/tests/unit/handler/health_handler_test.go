package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"service-short-link/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_GetServiceInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := handler.NewHealthHandler("1.2.3")
	r.GET("/", h.GetServiceInfo)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler.SuccessEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	m, _ := json.Marshal(resp.Data)
	var data map[string]interface{}
	json.Unmarshal(m, &data)

	assert.Equal(t, "ShortLink Service", data["name"])
	assert.Equal(t, "1.2.3", data["version"])
	assert.NotEmpty(t, data["uptime"])
	assert.NotEmpty(t, data["go_version"])
}

func TestHealthHandler_HealthAndProbes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := handler.NewHealthHandler("2.0.0")
	r.GET("/health", h.HealthCheck)
	r.GET("/ready", h.ReadinessProbe)
	r.GET("/live", h.LivenessProbe)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ready", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/live", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}
