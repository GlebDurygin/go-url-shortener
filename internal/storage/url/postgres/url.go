package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"
)

type URLPostgresStorage struct {
	database *sql.DB
}

func NewURLPostgresStorage(database *sql.DB) *URLPostgresStorage {
	return &URLPostgresStorage{database: database}
}

func (s *URLPostgresStorage) Save(originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		id := rand.Text()[:8]
		result, err := s.database.ExecContext(ctx,
			"INSERT INTO urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING",
			id, originalURL)

		if err != nil {
			return "", fmt.Errorf("save URL: %w", err)
		}

		count, err := result.RowsAffected()
		if err != nil {
			return "", fmt.Errorf("rows affected: %w", err)
		}

		if count == 1 {
			return id, nil
		}
	}
}

func (s *URLPostgresStorage) FindByID(id string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var originalURL string
	err := s.database.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", id).Scan(&originalURL)
	return originalURL, err == nil
}
