package file

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type URLFileStorage struct {
	mutex    sync.Mutex
	filePath string
}

type urlEntity struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURLFileStorage(path string) (*URLFileStorage, error) {
	m := &URLFileStorage{filePath: path}
	if _, err := m.read(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		if err := m.persist([]urlEntity{}); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *URLFileStorage) Save(originalURL string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entities, err := m.read()
	if err != nil {
		return "", err
	}

	for {
		id := rand.Text()[:8]
		duplicate := false
		for _, entity := range entities {
			if entity.ShortURL == id {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		entities = append(entities, urlEntity{UUID: rand.Text(), ShortURL: id, OriginalURL: originalURL})
		if err := m.persist(entities); err != nil {
			return "", err
		}
		return id, nil
	}
}

func (m *URLFileStorage) FindByID(id string) (string, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entities, err := m.read()
	if err != nil {
		return "", false
	}
	for _, entity := range entities {
		if entity.ShortURL == id {
			return entity.OriginalURL, true
		}
	}
	return "", false
}

func (m *URLFileStorage) persist(records []urlEntity) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode URL storage: %w", err)
	}
	if err := os.WriteFile(m.filePath, data, 0666); err != nil {
		return fmt.Errorf("write URL storage: %w", err)
	}
	return nil
}

func (m *URLFileStorage) read() ([]urlEntity, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("read URL storage: %w", err)
	}
	var entities []urlEntity
	if len(bytes.TrimSpace(data)) != 0 {
		if err := json.Unmarshal(data, &entities); err != nil {
			return nil, fmt.Errorf("decode URL storage: %w", err)
		}
	}
	return entities, nil
}
