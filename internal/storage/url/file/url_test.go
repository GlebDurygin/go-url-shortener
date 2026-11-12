package file

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLFileStorageSaveAndFindById(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)
	require.FileExists(t, path)
	url := "https://practicum.yandex.ru/"

	id, err := storage.Save(url)
	require.NoError(t, err)
	actual, ok := storage.FindByID(id)

	require.True(t, ok)
	assert.Equal(t, url, actual)
}

func TestURLFileStorageFindById(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	require.NoError(t, os.WriteFile(path, nil, 0600))
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)

	actual, ok := storage.FindByID("unknown_id")

	require.False(t, ok)
	assert.Empty(t, actual)
}

func TestURLFileStorageFindByIdUsesMemory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	data := []byte(`[{"uuid":"external","short_url":"known_id","original_url":"https://practicum.yandex.ru/"}]`)
	require.NoError(t, os.WriteFile(path, data, 0600))
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("invalid"), 0600))

	actual, ok := storage.FindByID("known_id")

	require.True(t, ok)
	assert.Equal(t, "https://practicum.yandex.ru/", actual)
}

func TestURLFileStorageSaveError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)
	require.NoError(t, os.Rename(path, path+".backup"))
	require.NoError(t, os.Mkdir(path, 0700))

	id, err := storage.Save("https://practicum.yandex.ru/")

	require.Error(t, err)
	assert.Empty(t, id)
	assert.Empty(t, storage.urls)
}

func TestURLFileStorageInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	require.NoError(t, os.WriteFile(path, []byte("invalid"), 0600))

	storage, err := NewURLFileStorage(path)

	require.Error(t, err)
	assert.Nil(t, storage)
}

func TestURLFileStorageBatchSaveAndFindById(t *testing.T) {
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
			path := filepath.Join(t.TempDir(), "urls.json")
			storage, err := NewURLFileStorage(path)
			require.NoError(t, err)
			existingID, err := storage.Save("https://existing.example")
			require.NoError(t, err)

			ids, err := storage.SaveAll(test.urls)

			require.NoError(t, err)
			require.Len(t, ids, len(test.urls))
			storage, err = NewURLFileStorage(path)
			require.NoError(t, err)
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

func TestURLFileStorageConcurrentSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)
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
	storage, err = NewURLFileStorage(path)
	require.NoError(t, err)
	assert.Len(t, storage.urls, 60)
}

func TestURLFileStorageSaveBatchError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")
	storage, err := NewURLFileStorage(path)
	require.NoError(t, err)
	require.NoError(t, os.Rename(path, path+".backup"))
	require.NoError(t, os.Mkdir(path, 0700))

	ids, err := storage.SaveAll([]string{"https://practicum.yandex.ru/", "https://example.org"})

	require.Error(t, err)
	assert.Empty(t, ids)
	assert.Empty(t, storage.urls)
}
