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
	return filepath.Join(home, ".commodo", "config.yaml")
}

func KeysPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".commodo", "keys.yaml.age")
}

func AgeKeyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".commodo", "age-key.txt")
}

// LoadKeys reads and decrypts the per-provider keys file.
// Returns empty map + nil error if the file does not exist (no keys saved yet).
// Returns an error if the file exists but cannot be decrypted.
func LoadKeys(path, ageKeyPath string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, nil
	}

	identity, err := loadAgeKey(ageKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load age key for decryption: %w", err)
	}

	plaintext, err := decryptKeys(data, identity)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt keys file: %w", err)
	}

	keys := map[string]string{}
	if err := yaml.Unmarshal(plaintext, &keys); err != nil {
		return nil, fmt.Errorf("cannot parse decrypted keys: %w", err)
	}
	return keys, nil
}

// SaveKeys encrypts and writes the per-provider keys file.
func SaveKeys(path, ageKeyPath string, keys map[string]string) error {
	plaintext, err := yaml.Marshal(keys)
	if err != nil {
		return fmt.Errorf("cannot marshal keys: %w", err)
	}

	identity, err := ensureAgeKey(ageKeyPath)
	if err != nil {
		return fmt.Errorf("cannot load age key for encryption: %w", err)
	}

	ciphertext, err := encryptKeys(plaintext, identity)
	if err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0600)
}

// Load reads config.yaml and resolves the API key from the encrypted keys file.
func Load(configPath, keysPath, ageKeyPath string) (*Config, error) {
	cfg, err := LoadFrom(configPath)
	if err != nil {
		return nil, err
	}

	if cfg.APIKey == "" {
		keys, err := LoadKeys(keysPath, ageKeyPath)
		if err != nil {
			return nil, err
		}
		cfg.APIKey = keys[cfg.Provider]
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("config: api_key not found (run commodo setup)")
	}

	return cfg, nil
}

// LoadFrom reads a config file. API key may be empty if stored in keys file.
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
	if cfg.Model == "" {
		cfg.Model = models.DefaultModel(cfg.Provider)
	}

	return &cfg, nil
}
