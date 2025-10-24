package database_test

import (
	domain "service-short-link/internal/domain"
	infra "service-short-link/internal/infrastructure/database"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := infra.NewAnalyticsRepository(db)
	row := &domain.Analytics{
		LinkID:    1,
		IPAddress: "127.0.0.1",
		UserAgent: ptr("test-agent"),
		Referrer:  ptr("test-referrer"),
		ClickedAt: time.Now(),
	}
	err := repo.Create(row)
	assert.NoError(t, err)
	assert.NotZero(t, row.ID)
}

func ptr[T any](v T) *T { return &v }
