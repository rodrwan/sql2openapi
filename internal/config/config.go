package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Security        SecurityConfig   `yaml:"security"`
	CustomEndpoints []CustomEndpoint `yaml:"customEndpoints"`
}

type SecurityConfig struct {
	Scheme    string          `yaml:"scheme"` // e.g. bearer
	Format    string          `yaml:"format"` // e.g. JWT
	Protected []ProtectedPath `yaml:"protected"`
}

type ProtectedPath struct {
	Path    string   `yaml:"path"`    // e.g. /users
	Methods []string `yaml:"methods"` // e.g. ["post", "put"]
}

type CustomEndpoint struct {
	Path           string                 `yaml:"path"`
	Method         string                 `yaml:"method"` // post, get, etc.
	Summary        string                 `yaml:"summary"`
	RequestSchema  map[string]interface{} `yaml:"requestSchema,omitempty"`
	ResponseSchema map[string]interface{} `yaml:"responseSchema,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error leyendo config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parseando config yaml: %w", err)
	}

	return &cfg, nil
}
