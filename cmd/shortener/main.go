package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/config"
	"github.com/GlebDurygin/go-url-shortener/internal/handler"
	"github.com/GlebDurygin/go-url-shortener/internal/router"
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

	engine := gin.New()
	shortenerHandler := handler.NewShortener(cfg.BaseURL, storage.NewURLStorage())
	router.RegisterRoutes(engine, shortenerHandler)

	err = http.ListenAndServe(cfg.ServerAddress, engine)
	if err != nil {
		return fmt.Errorf("listen and serve HTTP: %w", err)
	}
	return nil
}
