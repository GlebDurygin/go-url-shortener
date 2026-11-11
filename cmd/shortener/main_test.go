package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/file"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/inmemory"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/postgres"
	"github.com/stretchr/testify/require"
)

func TestNewURLStorage(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		database     *sql.DB
		expectedType any
	}{
		{
			name:         "database storage",
			filePath:     filepath.Join(t.TempDir(), "missing", "urls.json"),
			database:     &sql.DB{},
			expectedType: &postgres.URLPostgresStorage{},
		},
		{
			name:         "file storage",
			filePath:     filepath.Join(t.TempDir(), "urls.json"),
			expectedType: &file.URLFileStorage{},
		},
		{
			name:         "inmemory storage",
			expectedType: &inmemory.URLInMemoryStorage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := newURLStorage(tt.filePath, tt.database)
			require.NoError(t, err)
			require.IsType(t, tt.expectedType, storage)
		})
	}
}
