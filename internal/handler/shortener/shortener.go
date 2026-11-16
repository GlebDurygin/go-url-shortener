package shortener

import (
	"errors"
	"io"
	"net/http"
	"strings"

	dto "github.com/GlebDurygin/go-url-shortener/internal/dto/url"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/auth"
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

	id, err := s.repository.Save(body, c.GetString(auth.UserIDKey))
	if errors.Is(err, repository.ErrConflict) {
		c.Data(http.StatusConflict, "text/plain", []byte(s.baseURL+"/"+id))
		return
	} else if err != nil {
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

	id, err := s.repository.Save(body, c.GetString(auth.UserIDKey))
	if errors.Is(err, repository.ErrConflict) {
		s.writeJSON(c, http.StatusConflict, dto.CreateURLResponse{Result: s.baseURL + "/" + id})
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	s.writeJSON(c, http.StatusCreated, dto.CreateURLResponse{Result: s.baseURL + "/" + id})
}

func (s *Shortener) CreateShortenURLBatch(c *gin.Context) {
	request, err := s.validateURLBatch(c)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}

	urls := make([]string, len(request))
	for i, item := range request {
		urls[i] = item.OriginalURL
	}
	ids, err := s.repository.SaveAll(urls, c.GetString(auth.UserIDKey))
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	response := make(dto.CreateURLBatchResponses, len(request))
	for i, item := range request {
		response[i] = dto.CreateURLBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.baseURL + "/" + ids[i],
		}
	}
	s.writeJSON(c, http.StatusCreated, response)
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

func (s *Shortener) GetUserURLs(c *gin.Context) {
	userID := c.GetString(auth.UserIDKey)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	urls, err := s.repository.FindByUserID(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	if len(urls) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	response := make(dto.GetUserURLsResponse, 0, len(urls))
	for id, originalURL := range urls {
		response = append(response, dto.GetUserURLResponse{ShortURL: s.baseURL + "/" + id, OriginalURL: originalURL})
	}
	s.writeJSON(c, http.StatusOK, response)
}

func (s *Shortener) validateURL(c *gin.Context) (string, error) {
	body, err := s.readBody(c, "text/plain")
	if err != nil {
		return "", err
	}

	stringBody := strings.TrimSpace(string(body))
	if stringBody == "" {
		return "", errors.New("empty URL")
	}
	return stringBody, nil
}

func (s *Shortener) validateURLJson(c *gin.Context) (string, error) {
	body, err := s.readBody(c, "application/json")
	if err != nil {
		return "", err
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

func (s *Shortener) validateURLBatch(c *gin.Context) (dto.CreateURLBatchRequests, error) {
	body, err := s.readBody(c, "application/json")
	if err != nil {
		return nil, err
	}

	var request dto.CreateURLBatchRequests
	if err := easyjson.Unmarshal(body, &request); err != nil {
		return nil, err
	}

	if len(request) == 0 {
		return nil, errors.New("empty batch")
	}

	if err := binding.Validator.ValidateStruct(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *Shortener) readBody(c *gin.Context, contentType string) ([]byte, error) {
	if c.ContentType() != contentType {
		return nil, errors.New("invalid Content-Type")
	}
	return io.ReadAll(c.Request.Body)
}

func (s *Shortener) writeJSON(c *gin.Context, status int, response easyjson.Marshaler) {
	body, err := easyjson.Marshal(response)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}
	c.Data(status, "application/json", body)
}
