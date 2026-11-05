package file

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/GlebDurygin/go-url-shortener/internal/model"
)

type URLFileStorage struct {
	mutex    sync.Mutex
	filePath string
}

func NewURLFileStorage(path string) (*URLFileStorage, error) {
	m := &URLFileStorage{filePath: path}
	if _, err := m.read(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return m, nil
}

func (m *URLFileStorage) Save(originalURL string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	records, err := m.read()
	if err != nil {
		return "", err
	}

	for {
		id := rand.Text()[:8]
		duplicate := false
		for _, record := range records {
			if record.ShortURL == id {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		records = append(records, model.UrlEntity{UUID: rand.Text(), ShortURL: id, OriginalURL: originalURL})
		if err := m.persist(records); err != nil {
			return "", err
		}
		return id, nil
	}
}

func (m *URLFileStorage) FindByID(id string) (string, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	records, err := m.read()
	if err != nil {
		return "", false
	}
	for _, record := range records {
		if record.ShortURL == id {
			return record.OriginalURL, true
		}
	}
	return "", false
}

func (m *URLFileStorage) persist(records []model.UrlEntity) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode URL storage: %w", err)
	}
	if err := os.WriteFile(m.filePath, data, 0666); err != nil {
		return fmt.Errorf("write URL storage: %w", err)
	}
	return nil
}

func (m *URLFileStorage) read() ([]model.UrlEntity, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("read URL storage: %w", err)
	}
	var records []model.UrlEntity
	if len(bytes.TrimSpace(data)) != 0 {
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, fmt.Errorf("decode URL storage: %w", err)
		}
	}
	return records, nil
}
