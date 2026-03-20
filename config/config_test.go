package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/garciasdos/commodo/models"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: openai\nmodel: gpt-4o-mini\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.Provider)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", cfg.Model)
	}
}

func TestLoadDefaultModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: deepseek\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := models.DefaultModel("deepseek")
	if cfg.Model != expected {
		t.Errorf("expected default model %s, got %s", expected, cfg.Model)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := LoadFrom("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadMissingProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("model: gpt-4o-mini\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestLoadInvalidProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: fakellm\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}

func TestSaveAndLoadKeysRoundTrip(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	original := map[string]string{"openai": "sk-test123", "anthropic": "sk-ant456"}
	if err := SaveKeys(keysPath, ageKeyPath, original); err != nil {
		t.Fatalf("SaveKeys failed: %v", err)
	}

	loaded, err := LoadKeys(keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("LoadKeys failed: %v", err)
	}

	for k, v := range original {
		if loaded[k] != v {
			t.Errorf("key %s: expected %s, got %s", k, v, loaded[k])
		}
	}
}

func TestLoadKeysMissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "nonexistent.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	keys, err := LoadKeys(keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty map, got %v", keys)
	}
}

func TestLoadKeysExistButAgeKeyMissing(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	// Save keys (creates age-key.txt)
	SaveKeys(keysPath, ageKeyPath, map[string]string{"openai": "sk-test"})

	// Remove age key
	os.Remove(ageKeyPath)

	_, err := LoadKeys(keysPath, ageKeyPath)
	if err == nil {
		t.Fatal("expected error when age key is missing but keys file exists")
	}
}

func TestLoadResolvesKeyFromEncryptedKeysFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	os.WriteFile(configPath, []byte("provider: openai\nmodel: gpt-4o-mini\n"), 0644)
	SaveKeys(keysPath, ageKeyPath, map[string]string{"openai": "sk-from-keys"})

	cfg, err := Load(configPath, keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "sk-from-keys" {
		t.Errorf("expected api_key from encrypted keys, got %s", cfg.APIKey)
	}
}

func TestLoadMissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	os.WriteFile(configPath, []byte("provider: openai\n"), 0644)

	_, err := Load(configPath, keysPath, ageKeyPath)
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
}

func TestLoadFallsBackToInlineKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	// Config with api_key inline — should still work
	os.WriteFile(configPath, []byte("provider: openai\napi_key: sk-inline\nmodel: gpt-4o-mini\n"), 0644)

	cfg, err := Load(configPath, keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "sk-inline" {
		t.Errorf("expected inline api_key preserved, got %s", cfg.APIKey)
	}
}
