package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/handler"
)

func main() {
	mux := gin.New()
	shortenerHandler := handler.NewShortener()
	mux.POST("/", shortenerHandler.CreateShortenURL)
	mux.GET("/:id", shortenerHandler.RedirectToOriginalURL)
	mux.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
