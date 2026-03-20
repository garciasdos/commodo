package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garciasdos/commodo/models"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "commodo", "config.yaml")
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Provider == "" {
		return nil, fmt.Errorf("config: provider is required")
	}
	if !models.ValidProviders()[cfg.Provider] {
		return nil, fmt.Errorf("config: invalid provider %q (use: deepseek, openai, anthropic)", cfg.Provider)
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("config: api_key is required")
	}
	if cfg.Model == "" {
		cfg.Model = models.DefaultModel(cfg.Provider)
	}

	return &cfg, nil
}
