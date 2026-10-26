package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Parse(args []string) (Config, error) {
	var config Config
	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&config.ServerAddress, "a", ":8080", "Server HTTP address")
	flags.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	err := flags.Parse(args)
	return config, err
}
