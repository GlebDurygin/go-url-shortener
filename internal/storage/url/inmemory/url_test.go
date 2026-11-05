package inmemory

import (
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
