package handler

import (
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

type Shortener struct {
	mutex sync.RWMutex
	urls  map[string]string
}

func NewShortener() *Shortener {
	return &Shortener{
		urls: make(map[string]string),
	}
}

func (s *Shortener) CreateShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := s.validateURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := s.saveURL(body)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("http://" + r.Host + "/" + id))
}

func (s *Shortener) validateURL(r *http.Request) (string, error) {
	mediaType := r.Header.Get("Content-Type")
	if mediaType != "text/plain" {
		return "", errors.New("invalid Content-Type")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", errors.New(err.Error())
	}

	stringBody := strings.TrimSpace(string(body))
	if stringBody == "" {
		return "", errors.New("empty URL")
	}
	return stringBody, nil
}

func (s *Shortener) saveURL(originalURL string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for {
		id := rand.Text()[:8]
		if _, ok := s.urls[id]; ok {
			continue
		}

		s.urls[id] = originalURL
		return id
	}
}

func (s *Shortener) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "empty URL", http.StatusBadRequest)
		return
	}

	originalURL, ok := s.getURL(id)
	if !ok {
		http.Error(w, "unknown id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Shortener) getURL(id string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, ok := s.urls[id]
	return originalURL, ok
}
