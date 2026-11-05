package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShortenerHandler interface {
	CreateShortenURL(*gin.Context)
	CreateShortenURLJson(*gin.Context)
	RedirectToOriginalURL(*gin.Context)
}

func RegisterRoutes(engine *gin.Engine, shortenerHandler ShortenerHandler) {
	engine.POST("/", shortenerHandler.CreateShortenURL)
	engine.POST("/api/shorten", shortenerHandler.CreateShortenURLJson)
	engine.GET("/:id", shortenerHandler.RedirectToOriginalURL)
	engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})
}
