package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSetup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Simulate user input: provider=openai, api_key=sk-test123, model=(default)
	input := strings.NewReader("1\nsk-test123\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check config file was created
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "provider: openai") {
		t.Errorf("expected provider openai in config, got:\n%s", content)
	}
	if !strings.Contains(content, "api_key: sk-test123") {
		t.Errorf("expected api_key in config, got:\n%s", content)
	}
	if !strings.Contains(content, "model: gpt-4o-mini") {
		t.Errorf("expected default model in config, got:\n%s", content)
	}
}

func TestRunSetupDeepSeek(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	input := strings.NewReader("2\nsk-deep\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "provider: deepseek") {
		t.Errorf("expected provider deepseek, got:\n%s", content)
	}
	if !strings.Contains(content, "model: deepseek-chat") {
		t.Errorf("expected default model deepseek-chat, got:\n%s", content)
	}
}

func TestRunSetupAnthropic(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	input := strings.NewReader("3\nsk-ant\nclaude-haiku-4-5-20251001\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "provider: anthropic") {
		t.Errorf("expected provider anthropic, got:\n%s", content)
	}
	if !strings.Contains(content, "model: claude-haiku-4-5-20251001") {
		t.Errorf("expected custom model, got:\n%s", content)
	}
}

func TestRunSetupInvalidProvider(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Invalid choice "5", then valid "1"
	input := strings.NewReader("5\n1\nsk-test\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if !strings.Contains(string(data), "provider: openai") {
		t.Errorf("expected provider openai after retry, got:\n%s", string(data))
	}
}

func TestRunSetupEmptyAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Empty API key, then valid one
	input := strings.NewReader("1\n\nsk-real\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if !strings.Contains(string(data), "api_key: sk-real") {
		t.Errorf("expected api_key sk-real, got:\n%s", string(data))
	}
}

func TestRunSetupOutputMessages(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	input := strings.NewReader("1\nsk-test\n\n")
	var out bytes.Buffer

	Run(input, &out, configPath)

	output := out.String()
	if !strings.Contains(output, "Provider") {
		t.Error("expected provider prompt in output")
	}
	if !strings.Contains(output, "API key") {
		t.Error("expected API key prompt in output")
	}
}
