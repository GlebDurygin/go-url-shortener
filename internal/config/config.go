package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Parse(args []string) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&cfg.ServerAddress, "a", ":8080", "Server HTTP address")
	flags.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	err := flags.Parse(args)
	return cfg, err
}
