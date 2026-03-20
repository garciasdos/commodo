package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: openai\napi_key: sk-test123\nmodel: gpt-4o-mini\n")
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
	if cfg.APIKey != "sk-test123" {
		t.Errorf("expected api_key sk-test123, got %s", cfg.APIKey)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", cfg.Model)
	}
}

func TestLoadDefaultModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: deepseek\napi_key: sk-test\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "deepseek-chat" {
		t.Errorf("expected default model deepseek-chat, got %s", cfg.Model)
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
	content := []byte("api_key: sk-test\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestLoadMissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: openai\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
}

func TestLoadInvalidProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: fakellm\napi_key: sk-test\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}
