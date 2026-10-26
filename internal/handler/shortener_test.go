package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenerCreateShortenURL(t *testing.T) {
	type requestParams struct {
		contentType string
		body        string
		errorReader bool
		errorWriter bool
	}
	type expected struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestParams requestParams
		expected      expected
	}{
		{
			name:          "invalid content type",
			requestParams: requestParams{contentType: "application/json"},
			expected:      expected{statusCode: http.StatusBadRequest, body: "invalid Content-Type"},
		},
		{
			name:          "invalid body",
			requestParams: requestParams{contentType: "text/plain", errorReader: true},
			expected:      expected{statusCode: http.StatusBadRequest, body: "read failed"},
		},
		{
			name:          "empty URL",
			requestParams: requestParams{contentType: "text/plain", body: ""},
			expected:      expected{statusCode: http.StatusBadRequest, body: "empty URL"},
		},
		{
			name: "invalid writer",
			requestParams: requestParams{
				contentType: "text/plain",
				body:        "https://practicum.yandex.ru/",
				errorWriter: true,
			},
			expected: expected{statusCode: http.StatusCreated, body: ""},
		},
		{
			name:          "valid URL",
			requestParams: requestParams{contentType: "text/plain", body: "https://practicum.yandex.ru/"},
			expected:      expected{statusCode: http.StatusCreated},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reader io.Reader
			if test.requestParams.errorReader {
				reader = errorReader{}
			} else {
				reader = strings.NewReader(test.requestParams.body)
			}
			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", reader)
			r.Header.Set("Content-Type", test.requestParams.contentType)

			var w http.ResponseWriter
			var recorder = httptest.NewRecorder()
			if test.requestParams.errorWriter {
				w = &errorWriter{
					ResponseRecorder: recorder,
				}
			} else {
				w = recorder
			}

			s := NewShortener()
			c, _ := gin.CreateTestContext(w)
			c.Request = r
			s.CreateShortenURL(c)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.expected.statusCode, result.StatusCode)

			if test.expected.body != "" {
				resultBody, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.Equal(t, test.expected.body, string(resultBody))
			} else {
				assert.NotEmpty(t, result.Body)
			}
		})
	}
}

func TestShortenerRedirectToOriginalURL(t *testing.T) {
	type requestParams struct {
		id string
	}
	type expected struct {
		statusCode     int
		body           string
		locationHeader string
	}
	tests := []struct {
		name          string
		requestParams requestParams
		expected      expected
	}{
		{"empty id", requestParams{id: ""}, expected{statusCode: http.StatusBadRequest, body: "empty URL\n"}},
		{
			"unknown id",
			requestParams{id: "unknown_id"},
			expected{statusCode: http.StatusBadRequest, body: "unknown id\n"},
		},
		{
			"known id",
			requestParams{id: "known_id"},
			expected{statusCode: http.StatusTemporaryRedirect, locationHeader: "https://practicum.yandex.ru/"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
			var w = httptest.NewRecorder()

			s := NewShortener()
			s.urls["known_id"] = "https://practicum.yandex.ru/"
			c, _ := gin.CreateTestContext(w)
			c.Request = r
			c.Params = gin.Params{{Key: "id", Value: test.requestParams.id}}
			s.RedirectToOriginalURL(c)
			c.Writer.WriteHeaderNow()

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, test.expected.statusCode, result.StatusCode)
			assert.Equal(t, test.expected.locationHeader, result.Header.Get("Location"))
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

type errorWriter struct {
	*httptest.ResponseRecorder
}

func (w *errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
