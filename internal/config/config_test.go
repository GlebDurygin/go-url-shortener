package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{"parse defaults", nil, Config{":8080", "http://localhost:8080"}},
		{"parse server address", []string{"-a=:8888"}, Config{":8888", "http://localhost:8080"}},
		{"parse base URL", []string{"-b=https://yandex.practicum"}, Config{":8080", "https://yandex.practicum"}},
		{
			"parse all flags",
			[]string{"-a", ":8888", "-b", "http://localhost:9000"},
			Config{":8888", "http://localhost:9000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
