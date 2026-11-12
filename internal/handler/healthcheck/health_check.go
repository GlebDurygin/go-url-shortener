package healthcheck

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheck struct {
	database *sql.DB
}

func NewHealthCheck(database *sql.DB) *HealthCheck {
	return &HealthCheck{database: database}
}

func (h *HealthCheck) Ping(c *gin.Context) {
	if h.database == nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
	defer cancel()
	if err := h.database.PingContext(ctx); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
