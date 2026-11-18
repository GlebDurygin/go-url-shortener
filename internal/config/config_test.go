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
		authSecret      string
		args            []string
		expected        Config
	}{
		{
			"empty storage flags",
			"",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-f=", "-d="},
			Config{":8080", "http://localhost:8080", "", "", "test-env-secret"},
		},
		{
			"file environment overrides flag",
			"",
			"",
			"env.json",
			"",
			"test-env-secret",
			[]string{"-f=flag.json"},
			Config{":8080", "http://localhost:8080", "env.json", "", "test-env-secret"},
		},
		{
			"empty database flag with file environment",
			"",
			"",
			"env.json",
			"",
			"test-env-secret",
			[]string{"-d="},
			Config{":8080", "http://localhost:8080", "env.json", "", "test-env-secret"},
		},
		{
			"defaults",
			"",
			"",
			"",
			"",
			"test-env-secret",
			nil,
			Config{":8080", "http://localhost:8080", "", "", "test-env-secret"},
		},
		{
			"serverAddress flag",
			"",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-a=:8888"},
			Config{":8888", "http://localhost:8080", "", "", "test-env-secret"},
		},
		{
			"baseURL flag",
			"",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-b=https://practicum.yandex.ru/"},
			Config{":8080", "https://practicum.yandex.ru/", "", "", "test-env-secret"},
		},
		{
			"all flags",
			"",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-a", ":8888", "-b", "https://practicum.yandex.ru/"},
			Config{":8888", "https://practicum.yandex.ru/", "", "", "test-env-secret"},
		},
		{
			"environment overrides flags",
			":9000",
			"https://yandex.practicum",
			"",
			"",
			"test-env-secret",
			[]string{"-a=:8888", "-b=https://practicum.yandex.ru/"},
			Config{":9000", "https://yandex.practicum", "", "", "test-env-secret"},
		},
		{
			"serverAddress environment with baseURL default",
			":9000",
			"",
			"",
			"",
			"test-env-secret",
			nil,
			Config{":9000", "http://localhost:8080", "", "", "test-env-secret"},
		},
		{
			"baseURL environment with serverAddress default",
			"",
			"https://yandex.practicum",
			"",
			"",
			"test-env-secret",
			nil,
			Config{":8080", "https://yandex.practicum", "", "", "test-env-secret"},
		},
		{
			"serverAddress environment overrides flag",
			":9000",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-a=:8888", "-b=https://practicum.yandex.ru/"},
			Config{":9000", "https://practicum.yandex.ru/", "", "", "test-env-secret"},
		},
		{
			"baseURL environment overrides flag",
			"",
			"https://yandex.practicum",
			"",
			"",
			"test-env-secret",
			[]string{"-a=:8888", "-b=https://practicum.yandex.ru/"},
			Config{":8888", "https://yandex.practicum", "", "", "test-env-secret"},
		},
		{
			"empty flags",
			"",
			"",
			"",
			"",
			"test-env-secret",
			[]string{"-a=", "-b="},
			Config{"", "", "", "", "test-env-secret"},
		},
		{
			"environment with file flag",
			":9000",
			"https://yandex.practicum",
			"",
			"",
			"test-env-secret",
			[]string{"-f=urls.json"},
			Config{":9000", "https://yandex.practicum", "urls.json", "", "test-env-secret"},
		},
		{
			"databaseDSN flag", "", "", "", "", "test-env-secret",
			[]string{"-d", "host=localhost dbname=shortener"},
			Config{":8080", "http://localhost:8080", "", "host=localhost dbname=shortener", "test-env-secret"},
		},
		{
			"databaseDSN environment", "", "", "", "host=database dbname=shortener", "test-env-secret",
			nil,
			Config{":8080", "http://localhost:8080", "", "host=database dbname=shortener", "test-env-secret"},
		},
		{
			"databaseDSN environment overrides flag", "", "", "", "host=database", "test-env-secret",
			[]string{"-d", "host=localhost"},
			Config{":8080", "http://localhost:8080", "", "host=database", "test-env-secret"},
		},
		{
			"other environment values with databaseDSN flag",
			":9000",
			"https://yandex.practicum",
			"urls.json",
			"",
			"test-env-secret",
			[]string{"-d", "host=localhost"},
			Config{":9000", "https://yandex.practicum", "urls.json", "host=localhost", "test-env-secret"},
		},
		{
			"authSecret environment",
			"",
			"",
			"",
			"",
			"test-env-secret",
			nil,
			Config{":8080", "http://localhost:8080", "", "", "test-env-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", tt.fileStoragePath)
			t.Setenv("DATABASE_DSN", tt.databaseDSN)
			t.Setenv("AUTH_SECRET", tt.authSecret)

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
		{"missing authSecret value", "", "", []string{"-s"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", "")
			t.Setenv("DATABASE_DSN", "")
			t.Setenv("AUTH_SECRET", "")

			cfg, err := Parse(tt.args)
			require.Error(t, err)
			assert.Equal(t, Config{}, cfg)
		})
	}
}
