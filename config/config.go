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

func KeysPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "commodo", "keys.yaml")
}

func LoadKeys(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	keys := map[string]string{}
	yaml.Unmarshal(data, &keys)
	return keys
}

func SaveKeys(path string, keys map[string]string) error {
	data, err := yaml.Marshal(keys)
	if err != nil {
		return fmt.Errorf("cannot marshal keys: %w", err)
	}
	return os.WriteFile(path, data, 0600)
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
		return nil, fmt.Errorf("config: invalid provider %q (use: anthropic, deepseek, openai, openrouter)", cfg.Provider)
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("config: api_key is required")
	}
	if cfg.Model == "" {
		cfg.Model = models.DefaultModel(cfg.Provider)
	}

	return &cfg, nil
}
