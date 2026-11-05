package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

func RequestLogger() gin.HandlerFunc {
	return func(cfg *gin.Context) {
		start := time.Now()
		cfg.Next()

		Log.Info("HTTP request",
			zap.String("uri", cfg.Request.RequestURI),
			zap.String("method", cfg.Request.Method),
			zap.Int("status", cfg.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.Int("size", cfg.Writer.Size()),
		)
	}
}
