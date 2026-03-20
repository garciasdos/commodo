package models

import (
	_ "embed"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed models-pricing.yaml
var raw []byte

type Provider struct {
	Prefix   string   `yaml:"prefix"`
	Patterns []string `yaml:"patterns"`
	Default  string   `yaml:"default"`
}

type Config struct {
	Providers map[string]Provider `yaml:"providers"`
}

func Load() (*Config, error) {
	return Parse(raw)
}

func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("models: %w", err)
	}
	return &cfg, nil
}

func DefaultModel(provider string) string {
	cfg, err := Load()
	if err != nil {
		return ""
	}
	p, ok := cfg.Providers[provider]
	if !ok {
		return ""
	}
	return p.Default
}

func ValidProviders() map[string]bool {
	cfg, err := Load()
	if err != nil {
		return nil
	}
	m := make(map[string]bool, len(cfg.Providers))
	for name := range cfg.Providers {
		m[name] = true
	}
	return m
}

type ProviderInfo struct {
	Name         string
	DefaultModel string
}

func Providers() []ProviderInfo {
	cfg, err := Load()
	if err != nil {
		return nil
	}
	providers := make([]ProviderInfo, 0, len(cfg.Providers))
	for name, p := range cfg.Providers {
		providers = append(providers, ProviderInfo{Name: name, DefaultModel: p.Default})
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})
	return providers
}
