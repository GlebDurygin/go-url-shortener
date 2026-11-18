package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	repository "github.com/GlebDurygin/go-url-shortener/internal/repository/url"
)

type queryRowContext func(context.Context, string, ...any) *sql.Row

type URLPostgresStorage struct {
	database *sql.DB
}

func NewURLPostgresStorage(database *sql.DB) *URLPostgresStorage {
	return &URLPostgresStorage{database: database}
}

func (s *URLPostgresStorage) Save(originalURL, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.saveURL(ctx, s.database.QueryRowContext, originalURL, userID)
}

func (s *URLPostgresStorage) FindByID(id string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var originalURL string
	err := s.database.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", id).Scan(&originalURL)
	return originalURL, err == nil
}

func (s *URLPostgresStorage) SaveAll(urls []string, userID string) ([]string, error) {
	ids := make([]string, 0, len(urls))
	if len(urls) == 0 {
		return ids, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin URL batch: %w", err)
	}
	defer tx.Rollback()

	var batchError error
	for _, originalURL := range urls {
		id, err := s.saveURL(ctx, tx.QueryRowContext, originalURL, userID)
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				batchError = errors.Join(batchError, err)
			} else {
				return nil, fmt.Errorf("save URL batch: %w", err)
			}
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit URL batch: %w", err)
	}

	return ids, batchError
}

func (s *URLPostgresStorage) FindByUserID(userID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.database.QueryContext(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1 AND user_id <> ''", userID)
	if err != nil {
		return nil, fmt.Errorf("find user URLs: %w", err)
	}
	defer rows.Close()

	urls := make(map[string]string)
	for rows.Next() {
		var id, originalURL string
		if err := rows.Scan(&id, &originalURL); err != nil {
			return nil, fmt.Errorf("scan user URL: %w", err)
		}
		urls[id] = originalURL
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read user URLs: %w", err)
	}
	return urls, nil
}

func (s *URLPostgresStorage) saveURL(ctx context.Context, queryRow queryRowContext, originalURL, userID string) (string, error) {
	var id string
	var conflictedOriginalURL bool

	err := queryRow(ctx, `
		INSERT INTO urls (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url)
		DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_url, OLD.short_url IS NOT NULL`,
		rand.Text()[:8], originalURL, userID).Scan(&id, &conflictedOriginalURL)
	if err != nil {
		return "", fmt.Errorf("save URL: %w", err)
	}

	if conflictedOriginalURL {
		return id, repository.ErrConflict
	}
	return id, nil
}
