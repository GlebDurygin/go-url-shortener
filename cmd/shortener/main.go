package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GlebDurygin/go-url-shortener/internal/handler/healthcheck"
	"github.com/GlebDurygin/go-url-shortener/internal/handler/shortener"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/compress"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/logger"
	"github.com/GlebDurygin/go-url-shortener/internal/repository/url"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/file"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/GlebDurygin/go-url-shortener/internal/config"
	"github.com/GlebDurygin/go-url-shortener/internal/router"
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

	urlStorage, err := newURLStorage(cfg.FileStoragePath)
	if err != nil {
		return err
	}

	var database *sql.DB
	if cfg.DatabaseDSN != "" {
		database, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("initialize database: invalid PostgreSQL connection configuration")
		}
		defer database.Close()
	}

	engine := gin.New()
	engine.Use(logger.RequestLogger())
	engine.Use(compress.Gzip())
	shortenerHandler := shortener.NewShortener(cfg.BaseURL, urlStorage)
	router.RegisterRoutes(engine, shortenerHandler, healthcheck.NewHealthCheck(database))

	err = http.ListenAndServe(cfg.ServerAddress, engine)
	if err != nil {
		return fmt.Errorf("listen and serve HTTP: %w", err)
	}
	return nil
}

func newURLStorage(filePath string) (url.URLRepository, error) {
	fileStorage, err := file.NewURLFileStorage(filePath)
	if err != nil {
		return nil, fmt.Errorf("initialize file storage %q: %w", filePath, err)
	}
	return fileStorage, nil
}
