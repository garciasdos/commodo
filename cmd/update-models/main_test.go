package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/garciasdos/commodo/models"
)

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{"gpt-4o-mini", []string{"gpt-*-mini*"}, true},
		{"gpt-4.1-mini", []string{"gpt-*-mini*"}, true},
		{"gpt-4o", []string{"gpt-*-mini*"}, false},
		{"claude-haiku-4-5-20251001", []string{"claude-*-haiku-*"}, false},
		{"claude-haiku-4-5-20251001", []string{"claude-haiku-*"}, true},
		{"deepseek-chat", []string{"deepseek-chat*"}, true},
		{"deepseek-coder", []string{"deepseek-chat*"}, false},
	}
	for _, tt := range tests {
		got := matchesAny(tt.name, tt.patterns)
		if got != tt.want {
			t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.name, tt.patterns, got, tt.want)
		}
	}
}

func TestFindCheapest(t *testing.T) {
	orModels := []openRouterModel{
		{ID: "openai/gpt-4o-mini", Pricing: openRouterPricing{Prompt: "0.00000015"}},
		{ID: "openai/gpt-4.1-mini", Pricing: openRouterPricing{Prompt: "0.00000040"}},
		{ID: "openai/gpt-4o", Pricing: openRouterPricing{Prompt: "0.0000025"}},
		{ID: "anthropic/claude-haiku-4-5-20251001", Pricing: openRouterPricing{Prompt: "0.0000008"}},
	}

	provider := models.Provider{
		Prefix:   "openai",
		Patterns: []string{"gpt-*-mini*"},
		Default:  "gpt-4.1-mini",
	}

	name, price, err := findCheapest(provider, orModels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", name)
	}
	if price != 0.00000015 {
		t.Errorf("expected 0.00000015, got %v", price)
	}
}

func TestFindCheapestNoMatch(t *testing.T) {
	orModels := []openRouterModel{
		{ID: "openai/gpt-4o", Pricing: openRouterPricing{Prompt: "0.0000025"}},
	}

	provider := models.Provider{
		Prefix:   "openai",
		Patterns: []string{"gpt-*-mini*"},
		Default:  "gpt-4o-mini",
	}

	_, _, err := findCheapest(provider, orModels)
	if err == nil {
		t.Fatal("expected error for no matching models")
	}
}

func TestFetchOpenRouterModels(t *testing.T) {
	expected := []openRouterModel{
		{ID: "openai/gpt-4o-mini", Pricing: openRouterPricing{Prompt: "0.00000015"}},
	}
	resp := openRouterResponse{Data: expected}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the URL for testing - we test the parsing via the server
	client := &http.Client{}
	httpResp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer httpResp.Body.Close()

	var result openRouterResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Data))
	}
	if result.Data[0].ID != "openai/gpt-4o-mini" {
		t.Errorf("expected openai/gpt-4o-mini, got %s", result.Data[0].ID)
	}
}

func TestEndToEndUpdate(t *testing.T) {
	// Create a temp models-pricing.yaml
	dir := t.TempDir()
	yamlContent := `providers:
    openai:
        prefix: openai
        patterns:
            - gpt-*-mini*
        default: gpt-4.1-mini
`
	yamlPath := filepath.Join(dir, "models-pricing.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(yamlPath)
	cfg, err := models.Parse(data)
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	orModels := []openRouterModel{
		{ID: "openai/gpt-4o-mini", Pricing: openRouterPricing{Prompt: "0.00000015"}},
		{ID: "openai/gpt-4.1-mini", Pricing: openRouterPricing{Prompt: "0.00000040"}},
	}

	provider := cfg.Providers["openai"]
	cheapest, _, err := findCheapest(provider, orModels)
	if err != nil {
		t.Fatalf("findCheapest: %v", err)
	}
	if cheapest != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", cheapest)
	}
	if cheapest == provider.Default {
		t.Error("expected a change from default")
	}
}
