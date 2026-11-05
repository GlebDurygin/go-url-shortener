package shortener

import (
	"errors"
	"io"
	"net/http"
	"strings"

	dto "github.com/GlebDurygin/go-url-shortener/internal/dto/url"
	repository "github.com/GlebDurygin/go-url-shortener/internal/repository/url"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mailru/easyjson"
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

	id, err := s.repository.Save(body)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	c.Data(http.StatusCreated, "text/plain", []byte(s.baseURL+"/"+id))
}

func (s *Shortener) CreateShortenURLJson(c *gin.Context) {
	body, err := s.validateURLJson(c)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}

	id, err := s.repository.Save(body)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	response, err := easyjson.Marshal(dto.CreateURLResponse{Result: s.baseURL + "/" + id})
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	c.Data(http.StatusCreated, "application/json", response)
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

func (s *Shortener) validateURLJson(c *gin.Context) (string, error) {
	mediaType := c.GetHeader("Content-Type")
	if mediaType != "application/json" {
		return "", errors.New("invalid Content-Type")
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", errors.New(err.Error())
	}

	var request dto.CreateURLRequest
	if err := easyjson.Unmarshal(body, &request); err != nil {
		return "", err
	}

	if err := binding.Validator.ValidateStruct(&request); err != nil {
		return "", err
	}
	return request.URL, nil
}
