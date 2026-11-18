package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		serverAddress   string
		baseURL         string
		fileStoragePath string
		databaseDSN     string
		args            []string
		expected        Config
	}{
		{"empty storage flags", "", "", "", "", []string{"-f=", "-d="}, Config{":8080", "http://localhost:8080", "", ""}},
		{"file environment overrides flag", "", "", "env.json", "", []string{"-f=flag.json"}, Config{":8080", "http://localhost:8080", "env.json", ""}},
		{"empty database flag with file environment", "", "", "env.json", "", []string{"-d="}, Config{":8080", "http://localhost:8080", "env.json", ""}},
		{"defaults", "", "", "", "", nil, Config{":8080", "http://localhost:8080", "", ""}},
		{
			"serverAddress flag",
			"",
			"",
			"",
			"",
			[]string{"-a=:8888"},
			Config{":8888", "http://localhost:8080", "", ""},
		},
		{
			"baseURL flag",
			"",
			"",
			"",
			"",
			[]string{"-b=https://flag.example"},
			Config{":8080", "https://flag.example", "", ""},
		},
		{
			"all flags",
			"",
			"",
			"",
			"",
			[]string{"-a", ":8888", "-b", "https://flag.example"},
			Config{":8888", "https://flag.example", "", ""},
		},
		{
			"environment overrides flags",
			":9000",
			"https://yandex.practicum",
			"",
			"",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":9000", "https://yandex.practicum", "", ""},
		},
		{
			"serverAddress environment with baseURL default",
			":9000",
			"",
			"",
			"",
			nil,
			Config{":9000", "http://localhost:8080", "", ""},
		},
		{
			"baseURL environment with serverAddress default",
			"",
			"https://yandex.practicum",
			"",
			"",
			nil,
			Config{":8080", "https://yandex.practicum", "", ""},
		},
		{
			"serverAddress environment overrides flag",
			":9000",
			"",
			"",
			"",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":9000", "https://flag.example", "", ""},
		},
		{
			"baseURL environment overrides flag",
			"",
			"https://yandex.practicum",
			"",
			"",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":8888", "https://yandex.practicum", "", ""},
		},
		{"empty flags", "", "", "", "", []string{"-a=", "-b="}, Config{"", "", "", ""}},
		{
			"environment with file flag",
			":9000",
			"https://yandex.practicum",
			"",
			"",
			[]string{"-f=urls.json"},
			Config{":9000", "https://yandex.practicum", "urls.json", ""},
		},
		{
			"databaseDSN flag", "", "", "", "",
			[]string{"-d", "host=localhost dbname=shortener"},
			Config{":8080", "http://localhost:8080", "", "host=localhost dbname=shortener"},
		},
		{
			"databaseDSN environment", "", "", "", "host=database dbname=shortener",
			nil,
			Config{":8080", "http://localhost:8080", "", "host=database dbname=shortener"},
		},
		{
			"databaseDSN environment overrides flag", "", "", "", "host=database",
			[]string{"-d", "host=localhost"},
			Config{":8080", "http://localhost:8080", "", "host=database"},
		},
		{
			"other environment values with databaseDSN flag", ":9000", "https://yandex.practicum", "urls.json", "",
			[]string{"-d", "host=localhost"},
			Config{":9000", "https://yandex.practicum", "urls.json", "host=localhost"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", tt.fileStoragePath)
			t.Setenv("DATABASE_DSN", tt.databaseDSN)

			cfg, err := Parse(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		baseURL       string
		args          []string
	}{
		{"unknown flag after valid flag", "", "", []string{"-a=:8888", "-unknown"}},
		{"missing serverAddress value", "", "", []string{"-a"}},
		{"missing baseURL value", "", "", []string{"-b"}},
		{"missing databaseDSN value", "", "", []string{"-d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", "")
			t.Setenv("DATABASE_DSN", "")

			cfg, err := Parse(tt.args)
			require.Error(t, err)
			assert.Equal(t, Config{}, cfg)
		})
	}
}
