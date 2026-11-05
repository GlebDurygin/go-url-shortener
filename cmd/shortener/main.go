package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GlebDurygin/go-url-shortener/internal/middleware/compress"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/GlebDurygin/go-url-shortener/internal/config"
	"github.com/GlebDurygin/go-url-shortener/internal/handler"
	"github.com/GlebDurygin/go-url-shortener/internal/router"
	"github.com/GlebDurygin/go-url-shortener/internal/storage"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Printf("failed to initialize logger: %v", err)
		return
	}

	if err := run(); err != nil {
		logger.Log.Error("server stopped", zap.Error(err))
	}
}

func run() error {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	engine := gin.New()
	engine.Use(logger.RequestLogger())
	engine.Use(compress.Gzip())
	shortenerHandler := handler.NewShortener(cfg.BaseURL, storage.NewURLStorage())
	router.RegisterRoutes(engine, shortenerHandler)

	err = http.ListenAndServe(cfg.ServerAddress, engine)
	if err != nil {
		return fmt.Errorf("listen and serve HTTP: %w", err)
	}
	return nil
}
