package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/repository"
)

type Shortener struct {
	repository repository.URLRepository
	baseURL    string
}

func NewShortener(baseURL string, repository repository.URLRepository) *Shortener {
	return &Shortener{
		repository: repository,
		baseURL:    baseURL,
	}
}

func (s *Shortener) CreateShortenURL(c *gin.Context) {
	body, err := s.validateURL(c)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}

	id := s.repository.Save(body)

	c.Data(http.StatusCreated, "text/plain", []byte(s.baseURL+"/"+id))
}

func (s *Shortener) RedirectToOriginalURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.String(http.StatusBadRequest, "empty URL")
		return
	}

	originalURL, ok := s.repository.FindByID(id)
	if !ok {
		c.String(http.StatusBadRequest, "unknown id")
		return
	}

	c.Header("Location", originalURL)
	c.Status(http.StatusTemporaryRedirect)
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
