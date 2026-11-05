package inmemory

import (
	"crypto/rand"
	"sync"
)

type URLInMemoryStorage struct {
	mutex sync.RWMutex
	urls  map[string]string
}

func NewURLInMemoryStorage() *URLInMemoryStorage {
	return &URLInMemoryStorage{urls: make(map[string]string)}
}

func (m *URLInMemoryStorage) Save(originalURL string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for {
		id := rand.Text()[:8]
		if _, ok := m.urls[id]; ok {
			continue
		}
		m.urls[id] = originalURL
		return id, nil
	}
}

func (m *URLInMemoryStorage) FindByID(id string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	originalURL, ok := m.urls[id]
	return originalURL, ok
}
