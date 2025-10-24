package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
)

type linkRepository struct {
	db *sql.DB
}

// NewLinkRepository creates a new instance of link repository
func NewLinkRepository(db *sql.DB) domain.LinkRepository {
	return &linkRepository{db: db}
}

// Create inserts a new link into the database
func (r *linkRepository) Create(link *domain.Link) error {
	query := `
		INSERT INTO links (short_code, original_url, title, description, is_active, expires_at, platform, role, channel_code, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query,
		link.ShortCode,
		link.OriginalURL,
		link.Title,
		link.Description,
		link.IsActive,
		link.ExpiresAt,
		link.Platform,
		link.Role,
		link.ChannelCode,
		link.CreatedAt,
		link.UpdatedAt,
	)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.Create: failed to create link", "short_code="+link.ShortCode, "error_type=database_create_failed")
		return fmt.Errorf("failed to create link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.Create: failed to get last insert id", "short_code="+link.ShortCode, "error_type=database_last_insert_id_failed")
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	link.ID = uint64(id)
	return nil
}

// GetByShortCode retrieves a link by its short code (full data)
func (r *linkRepository) GetByShortCode(shortCode string) (*domain.Link, error) {
	query := `
		SELECT id, short_code, original_url, title, description, is_active, expires_at, 
		       platform, role, channel_code, created_at, updated_at, click_count, last_accessed_at
		FROM links 
		WHERE short_code = ?
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	link := &domain.Link{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&link.ID,
		&link.ShortCode,
		&link.OriginalURL,
		&link.Title,
		&link.Description,
		&link.IsActive,
		&link.ExpiresAt,
		&link.Platform,
		&link.Role,
		&link.ChannelCode,
		&link.CreatedAt,
		&link.UpdatedAt,
		&link.ClickCount,
		&link.LastAccessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrLinkNotFound
		}
		logger.ErrorWithCockroachSimple(err, "LinkRepository.GetByShortCode: failed to get link by short code", "short_code="+shortCode, "error_type=database_query_failed")
		return nil, fmt.Errorf("failed to get link by short code: %w", err)
	}

	return link, nil
}

// GetForRedirect retrieves minimal link data optimized for redirection
func (r *linkRepository) GetForRedirect(shortCode string) (*domain.LinkForRedirect, error) {
	query := `
		SELECT id, original_url, is_active, expires_at
		FROM links 
		WHERE short_code = ?
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	link := &domain.LinkForRedirect{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.IsActive,
		&link.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrLinkNotFound
		}
		logger.ErrorWithCockroachSimple(err, "LinkRepository.GetForRedirect: failed to get link for redirect", "short_code="+shortCode, "error_type=database_query_failed")
		return nil, fmt.Errorf("failed to get link for redirect: %w", err)
	}

	return link, nil
}

func (r *linkRepository) ExistsByShortCode(shortCode string) (bool, error) {
	query := `SELECT COUNT(*) FROM links WHERE short_code = ?`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&count)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.ExistsByShortCode: failed to check short code existence", "short_code="+shortCode, "error_type=database_query_failed")
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return count > 0, nil
}

// IncrementClickCount atomically increments the click count for a link
func (r *linkRepository) IncrementClickCount(linkID uint64) error {
	query := `UPDATE links SET click_count = click_count + 1 WHERE id = ?`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.IncrementClickCount: failed to increment click count", "link_id="+fmt.Sprintf("%d", linkID), "error_type=database_update_failed")
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.IncrementClickCount: failed to get rows affected", "link_id="+fmt.Sprintf("%d", linkID), "error_type=database_rows_affected_failed")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrLinkNotFound
	}

	return nil
}

// UpdateLastAccessed updates the last accessed timestamp for a link
func (r *linkRepository) UpdateLastAccessed(linkID uint64) error {
	query := `UPDATE links SET last_accessed_at = ? WHERE id = ?`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, linkID)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.UpdateLastAccessed: failed to update last accessed time", "link_id="+fmt.Sprintf("%d", linkID), "error_type=database_update_failed")
		return fmt.Errorf("failed to update last accessed time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkRepository.UpdateLastAccessed: failed to get rows affected", "link_id="+fmt.Sprintf("%d", linkID), "error_type=database_rows_affected_failed")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrLinkNotFound
	}

	return nil
}
