package main

import (
	"net/http"

	"github.com/GlebDurygin/go-url-shortener/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	shortenerHandler := handler.NewShortener()
	mux.HandleFunc("POST /{$}", shortenerHandler.CreateShortenURL)
	mux.HandleFunc("GET /{id}", shortenerHandler.RedirectToOriginalURL)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
