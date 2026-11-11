package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GlebDurygin/go-url-shortener/internal/handler/healthcheck"
	"github.com/GlebDurygin/go-url-shortener/internal/handler/shortener"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/compress"
	"github.com/GlebDurygin/go-url-shortener/internal/middleware/logger"
	"github.com/GlebDurygin/go-url-shortener/internal/repository/url"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/file"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/inmemory"
	"github.com/GlebDurygin/go-url-shortener/internal/storage/url/postgres"
	"github.com/GlebDurygin/go-url-shortener/migrations"
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

	database, err := newDatabase(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	if database != nil {
		defer database.Close()
	}

	urlStorage, err := newURLStorage(cfg.FileStoragePath, database)
	if err != nil {
		return err
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

func newURLStorage(filePath string, database *sql.DB) (url.URLRepository, error) {
	if database != nil {
		return postgres.NewURLPostgresStorage(database), nil
	}
	if filePath == "" {
		return inmemory.NewURLInMemoryStorage(), nil
	}
	fileStorage, err := file.NewURLFileStorage(filePath)
	if err != nil {
		return nil, fmt.Errorf("initialize file storage %q: %w", filePath, err)
	}
	return fileStorage, nil
}

func newDatabase(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, nil
	}

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL connection url")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	if err := migrations.Up(ctx, database); err != nil {
		database.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return database, nil
}
