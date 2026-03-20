package models

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(cfg.Providers))
	}
}

func TestDefaultModel(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"openai", "gpt-4o-mini"},
		{"anthropic", "claude-haiku-4-5-20251001"},
		{"deepseek", "deepseek-chat"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		got := DefaultModel(tt.provider)
		if got != tt.want {
			t.Errorf("DefaultModel(%q) = %q, want %q", tt.provider, got, tt.want)
		}
	}
}

func TestValidProviders(t *testing.T) {
	vp := ValidProviders()
	for _, name := range []string{"openai", "anthropic", "deepseek"} {
		if !vp[name] {
			t.Errorf("expected %q to be valid", name)
		}
	}
	if vp["unknown"] {
		t.Error("expected unknown to be invalid")
	}
}

func TestProviders(t *testing.T) {
	providers := Providers()
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(providers))
	}
	// Sorted alphabetically
	if providers[0].Name != "anthropic" {
		t.Errorf("expected first provider to be anthropic, got %s", providers[0].Name)
	}
}

func TestParse(t *testing.T) {
	data := []byte(`providers:
  test:
    prefix: "test"
    patterns: ["test-*"]
    default: test-model
`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Providers["test"].Default != "test-model" {
		t.Error("expected test-model as default")
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse([]byte(":::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
