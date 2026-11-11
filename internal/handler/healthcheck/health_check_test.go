package healthcheck

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pingError error
		missing   bool
		canceled  bool
		status    int
	}{
		{name: "connected", status: http.StatusOK},
		{name: "unavailable", pingError: errors.New("database unavailable"), status: http.StatusInternalServerError},
		{name: "not configured", missing: true, status: http.StatusInternalServerError},
		{name: "request canceled", canceled: true, status: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var db *sql.DB
			if !tt.missing {
				db = sql.OpenDB(testConnector{t: t, err: tt.pingError})
				defer db.Close()
			}
			engine := gin.New()
			engine.GET("/ping", NewHealthCheck(db).Ping)
			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			if tt.canceled {
				ctx, cancel := context.WithCancel(request.Context())
				cancel()
				request = request.WithContext(ctx)
			}
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			require.Equal(t, tt.status, response.Code)
			require.Empty(t, response.Body.String())
		})
	}
}

type testConnector struct {
	t   *testing.T
	err error
}

func (c testConnector) Connect(context.Context) (driver.Conn, error) { return testConnection{c}, nil }
func (c testConnector) Driver() driver.Driver                        { return nil }

type testConnection struct{ testConnector }

func (c testConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c testConnection) Close() error              { return nil }
func (c testConnection) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }
func (c testConnection) Ping(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	require.True(c.t, ok)
	require.Greater(c.t, time.Until(deadline), time.Duration(0))
	require.LessOrEqual(c.t, time.Until(deadline), 5*time.Second)
	return c.err
}
