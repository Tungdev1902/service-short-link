package database_test

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	domain "service-short-link/internal/domain"
	configinfra "service-short-link/internal/infrastructure/config"
	infra "service-short-link/internal/infrastructure/database"
	services "service-short-link/internal/infrastructure/services"

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

func TestLinkRepository_CRUD(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	// Cleanup database after test
	defer cleanupTestDB(t, db)

	repo := infra.NewLinkRepository(db)
	generator := services.NewShortCodeGenerator()
	code := ""
	for i := 0; i < 10; i++ {
		cand := generator.GenerateWithLength(7)
		exists, err := repo.ExistsByShortCode(cand)
		if err != nil {
			t.Fatalf("check code fail: %v", err)
		}
		if !exists {
			code = cand
			break
		}
	}
	if code == "" {
		t.Fatalf("Không tạo được short code unique sau 10 lần thử")
	}
	link := &domain.Link{
		ShortCode:   code,
		OriginalURL: "https://test.demo.com",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := repo.Create(link)
	assert.NoError(t, err)
	assert.NotZero(t, link.ID)

	res, err := repo.GetByShortCode(link.ShortCode)
	assert.NoError(t, err)
	assert.Equal(t, link.OriginalURL, res.OriginalURL)
	_, err = repo.GetForRedirect(link.ShortCode)
	assert.NoError(t, err)
	exists, err := repo.ExistsByShortCode(link.ShortCode)
	assert.NoError(t, err)
	assert.True(t, exists)

	err = repo.IncrementClickCount(link.ID)
	assert.NoError(t, err)
	err = repo.UpdateLastAccessed(link.ID)
	assert.NoError(t, err)
}
