package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
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
	mediaType := r.Header.Get("Content-Type")
	if mediaType != "text/plain" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := s.saveURL(string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("http://" + r.Host + "/" + id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Shortener) saveURL(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	id := hex.EncodeToString(hash[:])

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.urls[id] = originalURL

	return id
}

func (s *Shortener) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	originalURL, ok := s.getURL(id)
	if !ok {
		http.Error(w, "Unknown id", http.StatusBadRequest)
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
