package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/handler"
)

func RegisterRoutes(engine *gin.Engine, shortenerHandler *handler.Shortener) {
	engine.POST("/", shortenerHandler.CreateShortenURL)
	engine.POST("/api/shorten", shortenerHandler.CreateShortenURLJson)
	engine.GET("/:id", shortenerHandler.RedirectToOriginalURL)
	engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})
}
