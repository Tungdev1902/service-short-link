package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "service-short-link/internal/domain"
	handler "service-short-link/internal/handler"
	cacheinfra "service-short-link/internal/infrastructure/cache"
	configinfra "service-short-link/internal/infrastructure/config"
	services "service-short-link/internal/infrastructure/services"
	usecase "service-short-link/internal/usecase"

	infra "service-short-link/internal/infrastructure/database"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func newTestDB(t *testing.T) *sql.DB {
	// Use config service to get test database config
	configService := configinfra.NewConfigService()
	dbConfig := configService.GetTestDatabaseConfig()

	// If running from host (not Docker), use localhost with mapped port
	if dbConfig.Host == "mariadb_test" {
		// Check if we can resolve mariadb_test (Docker network) with timeout
		done := make(chan bool, 1)
		var err error

		go func() {
			_, err = net.LookupHost("mariadb_test")
			done <- true
		}()

		select {
		case <-done:
			if err != nil {
				// Running from host, use localhost with mapped port
				dbConfig.Host = "localhost"
				dbConfig.Port = 3316 // Mapped port from docker-compose
				t.Logf("Running from host, using localhost:3316 instead of mariadb_test:3306")
			}
		case <-time.After(1 * time.Second):
			// Timeout - assume running from host
			dbConfig.Host = "localhost"
			dbConfig.Port = 3316 // Mapped port from docker-compose
			t.Logf("DNS lookup timeout, assuming host mode, using localhost:3316")
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false&allowNativePasswords=true&multiStatements=false&interpolateParams=true&readTimeout=3s&writeTimeout=3s&timeout=3s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	t.Logf("Using test database: %s@%s:%d/%s",
		dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("cannot connect to test db: %v", err)
	}
	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	// Use TRUNCATE for faster cleanup (faster than DELETE)
	queries := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE analytics",
		"TRUNCATE TABLE links",
		"SET FOREIGN_KEY_CHECKS = 1",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning: failed to execute %s: %v", query, err)
		}
	}
}

func newRedisClient(t *testing.T) (*redis.Client, func()) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return client, func() { client.Close(); s.Close() }
}

func TestLinkHandler_Integration(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	redisClient, cleanup := newRedisClient(t)
	defer cleanup()

	// Cleanup database after test
	defer cleanupTestDB(t, db)
	linkRepo := infra.NewLinkRepository(db)
	linkCache := cacheinfra.NewLinkCache(redisClient)
	analyticsRepo := infra.NewAnalyticsRepository(db)
	codeGenerator := services.NewShortCodeGenerator()
	uniqueCodeService := services.NewUniqueCodeService()
	qrGenerator := services.NewQRCodeGenerator()
	urlValidator := services.NewURLValidator()
	userAgentParser := services.NewUserAgentParser()
	configService := configinfra.NewConfigService()
	uc := usecase.NewLinkUseCase(
		linkRepo,
		linkCache,
		analyticsRepo,
		codeGenerator,
		uniqueCodeService,
		qrGenerator,
		urlValidator,
		userAgentParser,
		configService,
	)
	linkH := handler.NewLinkHandler(uc)

	r := gin.New()
	r.POST("/api/v1/links", linkH.CreateLink)
	r.GET("/:code", linkH.RedirectLink)
	r.GET("/qr/:code", linkH.GenerateQRCode)

	// 1. Test tạo link mới
	createReq := domain.CreateLinkRequest{
		OriginalURL: "https://integration-test.example.com/" + time.Now().Format("150405"),
	}
	b, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Handle case where data might be nil
	data, ok := resp["data"].(map[string]interface{})
	if !ok || data == nil {
		t.Fatalf("Expected data field in response, got: %v", resp)
	}

	shortUrl, ok := data["short_url"].(string)
	if !ok {
		t.Fatalf("Expected short_url field in data, got: %v", data)
	}

	shortCode := shortUrl[strings.LastIndex(shortUrl, "/")+1:]

	// 2. Test redirect
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/"+shortCode, nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 302, w2.Code)

	// 3. Test QR code
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/qr/"+shortCode, nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	assert.Equal(t, "image/png", w3.Header().Get("Content-Type"))
}
