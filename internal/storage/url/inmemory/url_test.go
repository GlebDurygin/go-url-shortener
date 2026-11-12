package inmemory

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLInMemoryStorageSaveAndFindById(t *testing.T) {
	storage := NewURLInMemoryStorage()
	url := "https://practicum.yandex.ru/"

	id, err := storage.Save(url)
	require.NoError(t, err)
	actual, ok := storage.FindByID(id)

	require.True(t, ok)
	assert.Equal(t, url, actual)
}

func TestURLInMemoryStorageFindById(t *testing.T) {
	storage := NewURLInMemoryStorage()

	actual, ok := storage.FindByID("unknown_id")

	require.False(t, ok)
	assert.Empty(t, actual)
}

func TestURLInMemoryStorageBatchSaveAndFindById(t *testing.T) {
	tests := []struct {
		name string
		urls []string
	}{
		{"empty batch", nil},
		{"single URL", []string{"https://practicum.yandex.ru/"}},
		{"multiple URLs", []string{"https://practicum.yandex.ru/", "https://example.org"}},
		{"duplicate URLs", []string{"https://practicum.yandex.ru/", "https://practicum.yandex.ru/"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewURLInMemoryStorage()
			existingID, err := storage.Save("https://existing.example")
			require.NoError(t, err)

			ids, err := storage.SaveAll(test.urls)

			require.NoError(t, err)
			require.Len(t, ids, len(test.urls))

			actual, ok := storage.FindByID(existingID)
			require.True(t, ok)
			assert.Equal(t, "https://existing.example", actual)
			seen := make(map[string]bool)
			for i, id := range ids {
				assert.NotEmpty(t, id)
				assert.False(t, seen[id])
				seen[id] = true
				actual, ok := storage.FindByID(id)
				require.True(t, ok)
				assert.Equal(t, test.urls[i], actual)
			}
		})
	}
}

func TestURLInMemoryStorageConcurrentSave(t *testing.T) {
	storage := NewURLInMemoryStorage()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			id, err := storage.Save("https://practicum.yandex.ru/")
			assert.NoError(t, err)
			actual, ok := storage.FindByID(id)
			assert.True(t, ok)
			assert.Equal(t, "https://practicum.yandex.ru/", actual)
			ids, err := storage.SaveAll([]string{"https://example.org", "https://example.net"})
			assert.NoError(t, err)
			for i, id := range ids {
				actual, ok := storage.FindByID(id)
				assert.True(t, ok)
				assert.Equal(t, []string{"https://example.org", "https://example.net"}[i], actual)
			}
		})
	}
	wg.Wait()

	assert.Len(t, storage.urls, 60)
}
