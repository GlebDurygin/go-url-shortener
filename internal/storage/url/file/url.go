package file

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type URLFileStorage struct {
	mutex    sync.RWMutex
	filePath string
	urls     map[string]string
	users    map[string]string
}

type urlEntity struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
}

func NewURLFileStorage(path string) (*URLFileStorage, error) {
	m := &URLFileStorage{filePath: path, urls: make(map[string]string), users: make(map[string]string)}
	entities, err := m.read()
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		m.urls[entity.ShortURL] = entity.OriginalURL
		m.users[entity.ShortURL] = entity.UserID
	}
	return m, nil
}

func (m *URLFileStorage) Save(originalURL, userID string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for {
		id := rand.Text()[:8]
		if _, exists := m.urls[id]; exists {
			continue
		}

		if err := m.persist([]urlEntity{
			{
				UserID:      userID,
				UUID:        rand.Text(),
				ShortURL:    id,
				OriginalURL: originalURL,
			},
		}); err != nil {
			return "", err
		}
		m.urls[id] = originalURL
		m.users[id] = userID
		return id, nil
	}
}

func (m *URLFileStorage) FindByID(id string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	originalURL, ok := m.urls[id]
	return originalURL, ok
}

func (m *URLFileStorage) SaveAll(urls []string, userID string) ([]string, error) {
	ids := make([]string, 0, len(urls))
	if len(urls) == 0 {
		return ids, nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	records := make([]urlEntity, 0, len(urls))
	pending := make(map[string]string, len(urls))
	for _, originalURL := range urls {
		for {
			id := rand.Text()[:8]
			if _, exists := m.urls[id]; exists {
				continue
			}
			if _, exists := pending[id]; exists {
				continue
			}
			pending[id] = originalURL
			ids = append(ids, id)
			records = append(records, urlEntity{
				UserID:      userID,
				UUID:        rand.Text(),
				ShortURL:    id,
				OriginalURL: originalURL,
			})
			break
		}
	}
	if err := m.persist(records); err != nil {
		return nil, err
	}
	for id, originalURL := range pending {
		m.urls[id] = originalURL
		m.users[id] = userID
	}
	return ids, nil
}

func (m *URLFileStorage) FindByUserID(userID string) (map[string]string, error) {
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

func (m *URLFileStorage) persist(records []urlEntity) error {
	f, err := os.OpenFile(m.filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("open URL storage for append: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(records); err != nil {
		return fmt.Errorf("write URL storage: %w", err)
	}
	return nil
}

func (m *URLFileStorage) read() ([]urlEntity, error) {
	f, err := os.OpenFile(m.filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open URL storage: %w", err)
	}
	defer f.Close()

	var entities []urlEntity
	decoder := json.NewDecoder(f)
	for {
		var records []urlEntity
		if err := decoder.Decode(&records); err != nil {
			if errors.Is(err, io.EOF) {
				return entities, nil
			}
			return nil, fmt.Errorf("decode URL storage: %w", err)
		}
		entities = append(entities, records...)
	}
}
