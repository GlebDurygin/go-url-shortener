package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		baseURL       string
		args          []string
		expected      Config
	}{
		{"defaults", "", "", nil, Config{":8080", "http://localhost:8080", "urls.json"}},
		{"serverAddress flag", "", "", []string{"-a=:8888"}, Config{":8888", "http://localhost:8080", "urls.json"}},
		{
			"baseURL flag",
			"",
			"",
			[]string{"-b=https://flag.example"},
			Config{":8080", "https://flag.example", "urls.json"},
		},
		{
			"all flags",
			"",
			"",
			[]string{"-a", ":8888", "-b", "https://flag.example"},
			Config{":8888", "https://flag.example", "urls.json"},
		},
		{
			"environment overrides flags",
			":9000",
			"https://yandex.practicum",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":9000", "https://yandex.practicum", "urls.json"},
		},
		{
			"serverAddress environment with baseURL default",
			":9000",
			"",
			nil,
			Config{":9000", "http://localhost:8080", "urls.json"},
		},
		{
			"baseURL environment with serverAddress default",
			"",
			"https://yandex.practicum",
			nil,
			Config{":8080", "https://yandex.practicum", "urls.json"},
		},
		{
			"serverAddress environment overrides flag",
			":9000",
			"",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":9000", "https://flag.example", "urls.json"},
		},
		{
			"baseURL environment overrides flag",
			"",
			"https://yandex.practicum",
			[]string{"-a=:8888", "-b=https://flag.example"},
			Config{":8888", "https://yandex.practicum", "urls.json"},
		},
		{"empty flags", "", "", []string{"-a=", "-b="}, Config{"", "", "urls.json"}},
		{
			"environment with file flag",
			":9000",
			"https://yandex.practicum",
			[]string{"-f=urls.json"},
			Config{":9000", "https://yandex.practicum", "urls.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", "")

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.serverAddress)
			t.Setenv("BASE_URL", tt.baseURL)
			t.Setenv("FILE_STORAGE_PATH", "")

			cfg, err := Parse(tt.args)
			require.Error(t, err)
			assert.Equal(t, Config{}, cfg)
		})
	}
}
