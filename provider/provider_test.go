package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIGenerateCommitMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Bearer token")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected JSON content type")
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "feat(auth): add login endpoint"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAI("test-key", "gpt-4o-mini", server.URL)
	msg, err := p.GenerateCommitMessage("diff", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "feat(auth): add login endpoint" {
		t.Errorf("expected commit message, got: %s", msg)
	}
}

func TestOpenAIGenerateSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "A Go CLI tool for commit messages."}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAI("test-key", "gpt-4o-mini", server.URL)
	summary, err := p.GenerateSummary("# README\nA tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "A Go CLI tool for commit messages." {
		t.Errorf("unexpected summary: %s", summary)
	}
}

func TestDeepSeekGenerateCommitMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "fix(api): handle nil response"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewDeepSeek("test-key", "deepseek-chat", server.URL)
	msg, err := p.GenerateCommitMessage("diff", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "fix(api): handle nil response" {
		t.Errorf("expected commit message, got: %s", msg)
	}
}

func TestAnthropicGenerateCommitMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("expected x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("expected anthropic-version header")
		}
		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "refactor: extract helper function"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropic("test-key", "claude-sonnet-4-6-20250514", server.URL)
	msg, err := p.GenerateCommitMessage("diff", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "refactor: extract helper function" {
		t.Errorf("expected commit message, got: %s", msg)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	p := NewOpenAI("bad-key", "gpt-4o-mini", server.URL)
	_, err := p.GenerateCommitMessage("diff", "summary")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{"openai", "openai", false},
		{"deepseek", "deepseek", false},
		{"anthropic", "anthropic", false},
		{"invalid", "fakellm", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProvider(tt.provider, "key", "model")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider(%q) error = %v, wantErr %v", tt.provider, err, tt.wantErr)
			}
		})
	}
}
