package inmemory

import (
	"crypto/rand"
	"sync"
)

type URLInMemoryStorage struct {
	mutex sync.RWMutex
	urls  map[string]string
	users map[string]string
}

func NewURLInMemoryStorage() *URLInMemoryStorage {
	return &URLInMemoryStorage{urls: make(map[string]string), users: make(map[string]string)}
}

func (m *URLInMemoryStorage) Save(originalURL, userID string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for {
		id := rand.Text()[:8]
		if _, ok := m.urls[id]; ok {
			continue
		}
		m.urls[id] = originalURL
		m.users[id] = userID
		return id, nil
	}
}

func (m *URLInMemoryStorage) FindByID(id string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	originalURL, ok := m.urls[id]
	return originalURL, ok
}

func (m *URLInMemoryStorage) SaveAll(urls []string, userID string) ([]string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ids := make([]string, 0, len(urls))
	for _, originalURL := range urls {
		for {
			id := rand.Text()[:8]
			if _, exists := m.urls[id]; exists {
				continue
			}
			m.urls[id] = originalURL
			m.users[id] = userID
			ids = append(ids, id)
			break
		}
	}
	return ids, nil
}

func (m *URLInMemoryStorage) FindByUserID(userID string) (map[string]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	urls := make(map[string]string)
	for id, owner := range m.users {
		if userID != "" && owner == userID {
			urls[id] = m.urls[id]
		}
	}
	return urls, nil
}
