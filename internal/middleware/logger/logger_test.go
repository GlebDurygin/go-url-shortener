package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name   string
		method string
		status int
		body   string
	}{
		{"created status", http.MethodPost, http.StatusCreated, "created"},
		{"invalid request", http.MethodPost, http.StatusBadRequest, "bad request"},
		{"empty response", http.MethodGet, http.StatusOK, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			previousLog := Log
			Log = zap.New(core)
			t.Cleanup(func() { Log = previousLog })

			router := gin.New()
			router.Use(RequestLogger())
			router.Any("/", func(c *gin.Context) {
				c.String(tt.status, "%s", tt.body)
			})

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(tt.method, "/", nil))

			require.Equal(t, tt.status, response.Code)
			require.Equal(t, tt.body, response.Body.String())
			require.Equal(t, 1, logs.Len())
			requestLog := logs.All()[0]
			require.Equal(t, zap.InfoLevel, requestLog.Level)
			require.Equal(t, "HTTP request", requestLog.Message)
			logFields := requestLog.ContextMap()
			require.Equal(t, "/", logFields["uri"])
			require.Equal(t, tt.method, logFields["method"])
			require.Equal(t, int64(response.Code), logFields["status"])
			require.IsType(t, time.Duration(0), logFields["duration"])
			require.Equal(t, int64(len(tt.body)), logFields["size"])
		})
	}
}
