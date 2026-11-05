package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGzipResponse(t *testing.T) {
	for _, test := range []struct {
		name, contentType, accept string
		compressed                bool
	}{
		{"json", "application/json", "gzip", true},
		{"html", "text/html; charset=utf-8", "br, gzip", true},
		{"plain", "text/plain", "gzip", false},
		{"no header", "application/json", "", false},
		{"zero quality ignored", "application/json", "gzip;q=0, *;q=1", true},
		{"weighted", "application/json", "gzip;q=0.5", true},
		{"wildcard without gzip", "application/json", "*", false},
		{"other encoding", "application/json", "br", false},
		{"uppercase without lowercase gzip", "application/json", "GZIP; Q=0.5", false},
		{"invalid quality ignored", "application/json", "gzip;q=invalid", true},
		{"out of range quality ignored", "application/json", "gzip;q=2", true},
		{"wildcard with zero quality", "application/json", "*;q=1, gzip;q=0", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(Gzip())
			engine.GET("/", func(c *gin.Context) {
				c.Header("Content-Length", "7")
				c.Data(http.StatusCreated, test.contentType, []byte("payload"))
			})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Accept-Encoding", test.accept)
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			require.Equal(t, http.StatusCreated, response.Code)
			require.Contains(t, response.Header().Values("Vary"), "Accept-Encoding")
			var body io.Reader = response.Body
			if test.compressed {
				require.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
				require.Empty(t, response.Header().Get("Content-Length"))
				reader, err := gzip.NewReader(body)
				require.NoError(t, err)
				defer reader.Close()
				body = reader
			} else {
				require.Empty(t, response.Header().Get("Content-Encoding"))
			}
			data, err := io.ReadAll(body)
			require.NoError(t, err)
			require.Equal(t, "payload", string(data))
		})
	}
}

func TestGzipRequest(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	corrupt := bytes.Clone(compressed.Bytes())
	corrupt[len(corrupt)-1] ^= 0xff
	for _, test := range []struct {
		name     string
		encoding string
		data     []byte
		status   int
	}{
		{"valid", "gzip", compressed.Bytes(), http.StatusOK},
		{"uppercase encoding", "GZIP", compressed.Bytes(), http.StatusOK},
		{"encoding with spaces", "  gzip  ", compressed.Bytes(), http.StatusOK},
		{"invalid header", "gzip", []byte("invalid"), http.StatusBadRequest},
		{"invalid checksum", "gzip", corrupt, http.StatusBadRequest},
		{"truncated", "gzip", compressed.Bytes()[:compressed.Len()-4], http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(Gzip())
			engine.POST("/", func(c *gin.Context) {
				require.Empty(t, c.GetHeader("Content-Encoding"))
				require.EqualValues(t, -1, c.Request.ContentLength)
				data, err := io.ReadAll(c.Request.Body)
				if err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				require.Equal(t, "payload", string(data))
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(test.data))
			request.Header.Set("Content-Encoding", test.encoding)
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
		})
	}
}
