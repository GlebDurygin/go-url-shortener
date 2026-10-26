package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/GlebDurygin/go-url-shortener/internal/config"
	"github.com/GlebDurygin/go-url-shortener/internal/handler"
)

func main() {
	cnf, err := config.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	mux := gin.New()
	shortenerHandler := handler.NewShortener(cnf.BaseURL)
	mux.POST("/", shortenerHandler.CreateShortenURL)
	mux.GET("/:id", shortenerHandler.RedirectToOriginalURL)
	mux.NoRoute(func(c *gin.Context) {
		c.String(http.StatusBadRequest, "Invalid request")
	})

	err = http.ListenAndServe(cnf.ServerAddress, mux)
	if err != nil {
		panic(err)
	}
}
