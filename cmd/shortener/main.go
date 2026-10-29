package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/config"
	"github.com/GlebDurygin/go-url-shortener/internal/handler"
	"github.com/GlebDurygin/go-url-shortener/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
	}
}

func run() error {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	mux := gin.New()
	shortenerHandler := handler.NewShortener(cfg.BaseURL, storage.NewURLStorage())
	mux.POST("/", shortenerHandler.CreateShortenURL)
	mux.GET("/:id", shortenerHandler.RedirectToOriginalURL)
	mux.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})

	err = http.ListenAndServe(cfg.ServerAddress, mux)
	if err != nil {
		return fmt.Errorf("listen and serve HTTP: %w", err)
	}
	return nil
}
