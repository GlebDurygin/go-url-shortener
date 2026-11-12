//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/GlebDurygin/go-url-shortener/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresStorageSaveAndFindById(t *testing.T) {
	db := setupTestDatabase(t)
	storage := NewURLPostgresStorage(db)

	url := "https://practicum.yandex.ru/"
	id, err := storage.Save(url)
	require.NoError(t, err)

	actual, ok := storage.FindByID(id)
	require.True(t, ok)
	assert.Equal(t, url, actual)
}

func TestPostgresStorageSaveBatch(t *testing.T) {
	tests := []struct {
		name string
		urls []string
	}{
		{"empty batch", nil},
		{"multiple URLs", []string{"https://practicum.yandex.ru/", "https://example.org"}},
		{"duplicate URLs", []string{"https://practicum.yandex.ru/", "https://practicum.yandex.ru/"}},
	}
	database := setupTestDatabase(t)
	storage := NewURLPostgresStorage(database)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ids, err := storage.SaveAll(test.urls)

			require.NoError(t, err)
			require.Len(t, ids, len(test.urls))
			for i, id := range ids {
				actual, ok := storage.FindByID(id)
				require.True(t, ok)
				assert.Equal(t, test.urls[i], actual)
			}
		})
	}
}

func TestPostgresStorageSaveBatchRollback(t *testing.T) {
	database := setupTestDatabase(t)
	storage := NewURLPostgresStorage(database)

	ids, err := storage.SaveAll([]string{"https://practicum.yandex.ru/", "invalid\x00URL"})

	require.Error(t, err)
	assert.Empty(t, ids)
	var count int
	require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count))
	assert.Zero(t, count)
}

func setupTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	container, err := pgcontainer.Run(ctx, "postgres:18-alpine",
		pgcontainer.WithDatabase("shortener_test"),
		pgcontainer.WithUsername("postgres"),
		pgcontainer.WithPassword(""),
		testcontainers.WithEnv(map[string]string{"POSTGRES_HOST_AUTH_METHOD": "trust"}),
		pgcontainer.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	require.NoError(t, migrations.Up(ctx, db))

	return db
}
