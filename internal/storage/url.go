package storage

import (
	"crypto/rand"
	"sync"
)

type URLStorage struct {
	mutex sync.RWMutex
	urls  map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{urls: make(map[string]string)}
}

func (m *URLStorage) Save(originalURL string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for {
		id := rand.Text()[:8]
		if _, ok := m.urls[id]; ok {
			continue
		}
		m.urls[id] = originalURL
		return id
	}
}

func (m *URLStorage) FindByID(id string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	originalURL, ok := m.urls[id]
	return originalURL, ok
}
