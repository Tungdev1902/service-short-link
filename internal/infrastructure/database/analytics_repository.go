package database

import (
	"database/sql"
	"fmt"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"
)

type analyticsRepository struct {
	db *sql.DB
}

// NewAnalyticsRepository creates a new instance of analytics repository
func NewAnalyticsRepository(db *sql.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// Create inserts a new analytics record into the database
func (r *analyticsRepository) Create(analytics *domain.Analytics) error {
	query := `
		INSERT INTO analytics (link_id, ip_address, user_agent, referrer, device, os, browser, clicked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		analytics.LinkID,
		analytics.IPAddress,
		analytics.UserAgent,
		analytics.Referrer,
		analytics.Device,
		analytics.OS,
		analytics.Browser,
		analytics.ClickedAt,
	)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "AnalyticsRepository.Create: failed to create analytics record", "link_id="+fmt.Sprintf("%d", analytics.LinkID), "error_type=database_create_failed")
		return fmt.Errorf("failed to create analytics record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "AnalyticsRepository.Create: failed to get last insert id", "link_id="+fmt.Sprintf("%d", analytics.LinkID), "error_type=database_last_insert_id_failed")
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	analytics.ID = uint64(id)
	return nil
}
