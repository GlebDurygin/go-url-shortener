package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Parse(args []string) (Config, error) {
	config, ok := parseEnvs()
	if ok {
		return config, nil
	}

	flagConfig, err := parseFlags(args)
	if err != nil {
		return Config{}, err
	}

	config.ServerAddress = valueOrDefault(config.ServerAddress, flagConfig.ServerAddress)
	config.BaseURL = valueOrDefault(config.BaseURL, flagConfig.BaseURL)

	return config, nil
}

func parseEnvs() (Config, bool) {
	config := Config{
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		BaseURL:       os.Getenv("BASE_URL"),
	}
	return config, config.ServerAddress != "" && config.BaseURL != ""
}

func parseFlags(args []string) (Config, error) {
	var config Config

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&config.ServerAddress, "a", ":8080", "Server HTTP address")
	flags.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")

	err := flags.Parse(args)
	return config, err
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
