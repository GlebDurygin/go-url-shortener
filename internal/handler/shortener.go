package handler

import (
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Shortener struct {
	mutex   sync.RWMutex
	urls    map[string]string
	baseURL string
}

func NewShortener(baseURL string) *Shortener {
	return &Shortener{
		urls:    make(map[string]string),
		baseURL: baseURL,
	}
}

func (s *Shortener) CreateShortenURL(c *gin.Context) {
	body, err := s.validateURL(c)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}

	id := s.saveURL(body)

	c.Data(http.StatusCreated, "text/plain", []byte(s.baseURL+"/"+id))
}

func (s *Shortener) validateURL(c *gin.Context) (string, error) {
	mediaType := c.GetHeader("Content-Type")
	if mediaType != "text/plain" {
		return "", errors.New("invalid Content-Type")
	}

	body, err := io.ReadAll(c.Request.Body)
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

func (s *Shortener) RedirectToOriginalURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.String(http.StatusBadRequest, "empty URL")
		return
	}

	originalURL, ok := s.getURL(id)
	if !ok {
		c.String(http.StatusBadRequest, "unknown id")
		return
	}

	c.Header("Location", originalURL)
	c.Status(http.StatusTemporaryRedirect)
}

func (s *Shortener) getURL(id string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, ok := s.urls[id]
	return originalURL, ok
}
