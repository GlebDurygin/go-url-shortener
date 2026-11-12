package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShortenerHandler interface {
	CreateShortenURL(*gin.Context)
	CreateShortenURLJson(*gin.Context)
	RedirectToOriginalURL(*gin.Context)
	CreateShortenURLBatch(*gin.Context)
}

type HealthCheckHandler interface {
	Ping(*gin.Context)
}

func RegisterRoutes(engine *gin.Engine, shortenerHandler ShortenerHandler, healthCheckHandler HealthCheckHandler) {
	engine.POST("/", shortenerHandler.CreateShortenURL)
	engine.POST("/api/shorten", shortenerHandler.CreateShortenURLJson)
	engine.POST("/api/shorten/batch", shortenerHandler.CreateShortenURLBatch)
	engine.GET("/:id", shortenerHandler.RedirectToOriginalURL)

	engine.GET("/ping", healthCheckHandler.Ping)

	engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})
}
