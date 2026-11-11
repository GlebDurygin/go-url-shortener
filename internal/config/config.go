package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func Parse(args []string) (Config, error) {
	config := parseEnvs()

	flagConfig, err := parseFlags(args)
	if err != nil {
		return Config{}, err
	}

	config.ServerAddress = valueOrDefault(config.ServerAddress, flagConfig.ServerAddress)
	config.BaseURL = valueOrDefault(config.BaseURL, flagConfig.BaseURL)
	config.FileStoragePath = valueOrDefault(config.FileStoragePath, flagConfig.FileStoragePath)
	config.DatabaseDSN = valueOrDefault(config.DatabaseDSN, flagConfig.DatabaseDSN)

	return config, nil
}

func parseEnvs() Config {
	config := Config{
		ServerAddress:   os.Getenv("SERVER_ADDRESS"),
		BaseURL:         os.Getenv("BASE_URL"),
		FileStoragePath: os.Getenv("FILE_STORAGE_PATH"),
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
	}
	return config
}

func parseFlags(args []string) (Config, error) {
	var config Config

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&config.ServerAddress, "a", ":8080", "Server HTTP address")
	flags.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flags.StringVar(&config.FileStoragePath, "f", "urls.json", "Path to JSON URL storage")
	flags.StringVar(&config.DatabaseDSN, "d", "", "PostgreSQL connection string")

	err := flags.Parse(args)
	return config, err
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
