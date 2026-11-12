package shortener

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GlebDurygin/go-url-shortener/internal/handler/healthcheck"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/compress"
	"github.com/GlebDurygin/go-url-shortener/internal/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenerCreateShortenURL(t *testing.T) {
	type requestParams struct {
		contentType       string
		body              string
		errorReader       bool
		errorWriter       bool
		compressedRequest bool
	}
	type expected struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestParams requestParams
		saveError     error
		expected      expected
	}{
		{
			name:          "save error",
			requestParams: requestParams{contentType: "text/plain", body: `https://practicum.yandex.ru/`},
			saveError:     errors.New("save failed"),
			expected:      expected{statusCode: http.StatusInternalServerError, body: "save failed"},
		},
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
		{
			name: "gzip request",
			requestParams: requestParams{
				contentType:       "text/plain",
				body:              `https://practicum.yandex.ru/`,
				compressedRequest: true,
			},
			expected: expected{statusCode: http.StatusCreated, body: `http://localhost:8080/known_id`},
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
			if test.requestParams.compressedRequest {
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				_, err := writer.Write([]byte(test.requestParams.body))
				require.NoError(t, err)
				require.NoError(t, writer.Close())
				reader = &compressed
			}

			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", reader)
			r.Header.Set("Content-Type", test.requestParams.contentType)
			if test.requestParams.compressedRequest {
				r.Header.Set("Accept-Encoding", "gzip")
				r.Header.Set("Content-Encoding", "gzip")
			}

			var w http.ResponseWriter
			var recorder = httptest.NewRecorder()
			if test.requestParams.errorWriter {
				w = &errorWriter{
					ResponseRecorder: recorder,
				}
			} else {
				w = recorder
			}

			repository := &stubURLRepository{err: test.saveError}
			s := NewShortener("http://localhost:8080", repository)
			engine := gin.New()
			engine.Use(compress.Gzip())
			router.RegisterRoutes(engine, s, healthcheck.NewHealthCheck(nil))
			engine.ServeHTTP(w, r)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.expected.statusCode, result.StatusCode)
			require.Empty(t, result.Header.Get("Content-Encoding"))

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

func TestShortenerCreateShortenURLJson(t *testing.T) {
	type requestParams struct {
		contentType       string
		body              string
		errorReader       bool
		errorWriter       bool
		compressedRequest bool
	}
	type expected struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestParams requestParams
		saveError     error
		expected      expected
	}{
		{
			name: "save error",
			requestParams: requestParams{
				contentType: "application/json",
				body:        `{"url":"https://practicum.yandex.ru/"}`,
			},
			saveError: errors.New("save failed"),
			expected:  expected{statusCode: http.StatusInternalServerError, body: "save failed"},
		},
		{
			name:          "invalid content type",
			requestParams: requestParams{contentType: "text/plain"},
			expected:      expected{statusCode: http.StatusBadRequest, body: "invalid Content-Type"},
		},
		{
			name:          "invalid body",
			requestParams: requestParams{contentType: "application/json", errorReader: true},
			expected:      expected{statusCode: http.StatusBadRequest, body: "read failed"},
		},
		{
			name:          "empty URL",
			requestParams: requestParams{contentType: "application/json", body: `{"url":""}`},
			expected:      expected{statusCode: http.StatusBadRequest, body: "required"},
		},
		{
			name:          "missing URL",
			requestParams: requestParams{contentType: "application/json", body: `{}`},
			expected:      expected{statusCode: http.StatusBadRequest, body: "required"},
		},
		{
			name:          "null URL",
			requestParams: requestParams{contentType: "application/json", body: `{"url":null}`},
			expected:      expected{statusCode: http.StatusBadRequest, body: "required"},
		},
		{
			name:          "non-string URL",
			requestParams: requestParams{contentType: "application/json", body: `{"url":42}`},
			expected:      expected{statusCode: http.StatusBadRequest, body: ""},
		},
		{
			name: "invalid writer",
			requestParams: requestParams{
				contentType: "application/json",
				body:        `{"url":"https://practicum.yandex.ru/"}`,
				errorWriter: true,
			},
			expected: expected{statusCode: http.StatusCreated, body: ""},
		},
		{
			name: "valid URL",
			requestParams: requestParams{
				contentType: "application/json",
				body:        `{"url":"https://practicum.yandex.ru/"}`,
			},
			expected: expected{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://localhost:8080/known_id"}`,
			},
		},
		{
			name: "gzip request",
			requestParams: requestParams{
				contentType:       "application/json",
				body:              `{"url":"https://practicum.yandex.ru/"}`,
				compressedRequest: true,
			},
			expected: expected{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://localhost:8080/known_id"}`,
			},
		},
		{
			name: "gzip response",
			requestParams: requestParams{
				compressedRequest: true,
				contentType:       "application/json",
				body:              `{"url":"https://practicum.yandex.ru/"}`,
			},
			expected: expected{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://localhost:8080/known_id"}`,
			},
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
			if test.requestParams.compressedRequest {
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				_, err := writer.Write([]byte(test.requestParams.body))
				require.NoError(t, err)
				require.NoError(t, writer.Close())
				reader = &compressed
			}

			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", reader)
			r.Header.Set("Content-Type", test.requestParams.contentType)
			if test.requestParams.compressedRequest {
				r.Header.Set("Accept-Encoding", "gzip")
				r.Header.Set("Content-Encoding", "gzip")
			}

			var w http.ResponseWriter
			var recorder = httptest.NewRecorder()
			if test.requestParams.errorWriter {
				w = &errorWriter{
					ResponseRecorder: recorder,
				}
			} else {
				w = recorder
			}

			repository := &stubURLRepository{err: test.saveError}
			s := NewShortener("http://localhost:8080", repository)
			engine := gin.New()
			engine.Use(compress.Gzip())
			router.RegisterRoutes(engine, s, healthcheck.NewHealthCheck(nil))
			engine.ServeHTTP(w, r)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.expected.statusCode, result.StatusCode)

			var responseBody io.Reader = result.Body
			if test.requestParams.compressedRequest {
				require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
				reader, err := gzip.NewReader(result.Body)
				require.NoError(t, err)
				defer reader.Close()
				responseBody = reader
			} else {
				require.Empty(t, result.Header.Get("Content-Encoding"))
			}

			resultBody, err := io.ReadAll(responseBody)
			require.NoError(t, err)
			if test.expected.statusCode == http.StatusCreated {
				assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
				if test.requestParams.errorWriter {
					assert.Empty(t, resultBody)
				} else {
					assert.JSONEq(t, test.expected.body, string(resultBody))
				}
			} else {
				assert.NotEmpty(t, resultBody)
				if test.expected.body != "" {
					assert.Contains(t, string(resultBody), test.expected.body)
				}
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

			s := NewShortener("http://localhost:8080", &stubURLRepository{})
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

type stubURLRepository struct {
	err   error
	batch []string
}

func (s *stubURLRepository) Save(string) (string, error) {
	return "known_id", s.err
}

func (s *stubURLRepository) FindByID(id string) (string, bool) {
	if id == "known_id" {
		return "https://practicum.yandex.ru/", true
	}
	return "", false
}

func (s *stubURLRepository) SaveAll(urls []string) ([]string, error) {
	s.batch = urls
	ids := make([]string, len(urls))
	for i := range urls {
		ids[i] = "known_id"
	}
	return ids, s.err
}

func TestShortenerCreateShortenURLBatch(t *testing.T) {
	type requestParams struct {
		contentType string
		body        string
		compressed  bool
	}
	tests := []struct {
		name          string
		requestParams requestParams
		saveError     error
		statusCode    int
		expectedURLs  []string
	}{
		{
			"valid batch",
			requestParams{
				"application/json",
				`[{"correlation_id":"first","original_url":"https://practicum.yandex.ru/"},{"correlation_id":"second","original_url":"https://example.org"}]`,
				false,
			},
			nil,
			http.StatusCreated,
			[]string{"https://practicum.yandex.ru/", "https://example.org"},
		},
		{
			"gzip batch",
			requestParams{
				"application/json; charset=utf-8",
				`[{"correlation_id":"first","original_url":"https://practicum.yandex.ru/"},{"correlation_id":"second","original_url":"https://example.org"}]`,
				true,
			},
			nil,
			http.StatusCreated,
			[]string{"https://practicum.yandex.ru/", "https://example.org"},
		},
		{"empty batch", requestParams{"application/json", "[]", false}, nil, http.StatusBadRequest, nil},
		{"null batch", requestParams{"application/json", "null", false}, nil, http.StatusBadRequest, nil},
		{"invalid JSON", requestParams{"application/json", "[", false}, nil, http.StatusBadRequest, nil},
		{"object instead of array", requestParams{"application/json", "{}", false}, nil, http.StatusBadRequest, nil},
		{"trailing JSON", requestParams{"application/json", "[] []", false}, nil, http.StatusBadRequest, nil},
		{"invalid content type", requestParams{"text/plain", "[]", false}, nil, http.StatusBadRequest, nil},
		{
			"missing correlation",
			requestParams{"application/json", `[{"original_url":"https://practicum.yandex.ru/"}]`, false},
			nil,
			http.StatusBadRequest,
			nil,
		},
		{
			"invalid second URL",
			requestParams{
				"application/json",
				`[{"correlation_id":"first","original_url":"https://practicum.yandex.ru/"},{"correlation_id":"second","original_url":""}]`,
				false,
			},
			nil,
			http.StatusBadRequest,
			nil,
		},
		{
			"save error",
			requestParams{
				"application/json",
				`[{"correlation_id":"first","original_url":"https://practicum.yandex.ru/"}]`,
				false,
			},
			errors.New("save failed"),
			http.StatusInternalServerError,
			[]string{"https://practicum.yandex.ru/"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(test.requestParams.body)
			if test.requestParams.compressed {
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				_, err := writer.Write([]byte(test.requestParams.body))
				require.NoError(t, err)
				require.NoError(t, writer.Close())
				reader = &compressed
			}
			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", reader)
			r.Header.Set("Content-Type", test.requestParams.contentType)
			if test.requestParams.compressed {
				r.Header.Set("Content-Encoding", "gzip")
				r.Header.Set("Accept-Encoding", "gzip")
			}
			repository := &stubURLRepository{err: test.saveError}
			engine := gin.New()
			engine.Use(compress.Gzip())
			router.RegisterRoutes(engine, NewShortener("http://localhost:8080", repository), healthcheck.NewHealthCheck(nil))
			recorder := httptest.NewRecorder()

			engine.ServeHTTP(recorder, r)

			result := recorder.Result()
			defer result.Body.Close()
			assert.Equal(t, test.statusCode, result.StatusCode)
			assert.Equal(t, test.expectedURLs, repository.batch)
			var responseBody io.Reader = result.Body
			if test.requestParams.compressed {
				require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
				reader, err := gzip.NewReader(result.Body)
				require.NoError(t, err)
				defer reader.Close()
				responseBody = reader
			}
			body, err := io.ReadAll(responseBody)
			require.NoError(t, err)
			if test.statusCode == http.StatusCreated {
				assert.Contains(t, result.Header.Get("Content-Type"), "application/json")
				assert.JSONEq(t, `[{"correlation_id":"first","short_url":"http://localhost:8080/known_id"},{"correlation_id":"second","short_url":"http://localhost:8080/known_id"}]`, string(body))
			}
		})
	}
}
