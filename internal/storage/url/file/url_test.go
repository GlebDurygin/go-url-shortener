package file

import (
	"os"
	"path/filepath"
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
