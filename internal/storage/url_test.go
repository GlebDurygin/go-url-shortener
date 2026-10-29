package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLStorageSaveAndFindById(t *testing.T) {
	storage := NewURLStorage()
	url := "https://practicum.yandex.ru/"

	id := storage.Save(url)
	actual, ok := storage.FindByID(id)

	require.True(t, ok)
	assert.Equal(t, url, actual)
}

func TestURLStorageFindById(t *testing.T) {
	storage := NewURLStorage()

	actual, ok := storage.FindByID("unknown_id")

	require.False(t, ok)
	assert.Empty(t, actual)
}
