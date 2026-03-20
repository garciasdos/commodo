package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/garciasdos/commodo/config"
	"github.com/garciasdos/commodo/models"
)

// models.Providers() returns alphabetically: anthropic(1), deepseek(2), openai(3), openrouter(4)

func keysPath(dir string) string {
	return filepath.Join(dir, "keys.yaml.age")
}

func ageKeyPath(dir string) string {
	return filepath.Join(dir, "age-key.txt")
}

func TestRunSetup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Choose openai (3rd alphabetically), api_key, default model
	input := strings.NewReader("3\nsk-test123\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "provider: openai") {
		t.Errorf("expected provider openai in config, got:\n%s", content)
	}
	if strings.Contains(content, "api_key") {
		t.Errorf("config.yaml should not contain api_key, got:\n%s", content)
	}
	expectedModel := models.DefaultModel("openai")
	if !strings.Contains(content, "model: "+expectedModel) {
		t.Errorf("expected default model %s in config, got:\n%s", expectedModel, content)
	}

	// Key should be in encrypted keys file
	keys, err2 := config.LoadKeys(keysPath(dir), ageKeyPath(dir))
	if err2 != nil {
		t.Fatalf("LoadKeys failed: %v", err2)
	}
	if keys["openai"] != "sk-test123" {
		t.Errorf("expected key sk-test123 in keys.yaml, got %s", keys["openai"])
	}
}

func TestRunSetupDeepSeek(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// deepseek is 2nd alphabetically
	input := strings.NewReader("2\nsk-deep\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "provider: deepseek") {
		t.Errorf("expected provider deepseek, got:\n%s", content)
	}
	expectedModel := models.DefaultModel("deepseek")
	if !strings.Contains(content, "model: "+expectedModel) {
		t.Errorf("expected default model %s, got:\n%s", expectedModel, content)
	}
}

func TestRunSetupAnthropic(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// anthropic is 1st alphabetically
	input := strings.NewReader("1\nsk-ant\nclaude-haiku-4-5-20251001\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
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

	// Invalid choice "6", then valid "1" (anthropic)
	input := strings.NewReader("6\n1\nsk-test\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if !strings.Contains(string(data), "provider: anthropic") {
		t.Errorf("expected provider anthropic after retry, got:\n%s", string(data))
	}
}

func TestRunSetupEmptyAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Empty API key, then valid one; choice 1 = anthropic
	input := strings.NewReader("1\n\nsk-real\n\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keys, err2 := config.LoadKeys(keysPath(dir), ageKeyPath(dir))
	if err2 != nil {
		t.Fatalf("LoadKeys failed: %v", err2)
	}
	if keys["anthropic"] != "sk-real" {
		t.Errorf("expected key sk-real in keys.yaml, got %s", keys["anthropic"])
	}
}

func TestRunSetupOutputMessages(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// choice 1 = anthropic
	input := strings.NewReader("1\nsk-test\n\n")
	var out bytes.Buffer

	Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)

	output := out.String()
	if !strings.Contains(output, "Provider") {
		t.Error("expected provider prompt in output")
	}
	if !strings.Contains(output, "API key") {
		t.Error("expected API key prompt in output")
	}
}

func TestRunSetupPersistsKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)
	akp := ageKeyPath(dir)

	// First setup: openai with key
	Run(strings.NewReader("3\nsk-openai-key\n\n"), &bytes.Buffer{}, configPath, kp, akp, false, false)

	// Second setup: switch to anthropic
	Run(strings.NewReader("1\nsk-ant-key\n\n"), &bytes.Buffer{}, configPath, kp, akp, false, false)

	// Third setup: switch back to openai, press enter to reuse saved key
	var out bytes.Buffer
	Run(strings.NewReader("3\n\n\n"), &out, configPath, kp, akp, false, false)

	keys, err := config.LoadKeys(kp, akp)
	if err != nil {
		t.Fatalf("LoadKeys failed: %v", err)
	}
	if keys["openai"] != "sk-openai-key" {
		t.Errorf("expected saved openai key to be reused, got %s", keys["openai"])
	}
	// Prompt should show masked key
	if !strings.Contains(out.String(), "sk-o***") {
		t.Errorf("expected masked key in prompt, got:\n%s", out.String())
	}
}

func TestRunSetupModelOnly(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)
	akp := ageKeyPath(dir)

	// Initial setup
	Run(strings.NewReader("3\nsk-openai-key\n\n"), &bytes.Buffer{}, configPath, kp, akp, false, false)

	// Model-only update
	var out bytes.Buffer
	err := Run(strings.NewReader("my-custom-model\n"), &out, configPath, kp, akp, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "model: my-custom-model") {
		t.Errorf("expected model updated, got:\n%s", content)
	}
	if !strings.Contains(content, "provider: openai") {
		t.Errorf("expected provider preserved, got:\n%s", content)
	}
	if strings.Contains(content, "api_key") {
		t.Errorf("config.yaml should not contain api_key, got:\n%s", content)
	}
}

func TestRunSetupFreeMode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)
	akp := ageKeyPath(dir)

	// Free mode: only prompts for API key; provider and model are pre-set
	input := strings.NewReader("sk-or-freekey\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, kp, akp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "provider: openrouter") {
		t.Errorf("expected provider openrouter, got:\n%s", content)
	}
	if strings.Contains(content, "api_key") {
		t.Errorf("config.yaml should not contain api_key, got:\n%s", content)
	}
	keys, err2 := config.LoadKeys(kp, akp)
	if err2 != nil {
		t.Fatalf("LoadKeys failed: %v", err2)
	}
	if keys["openrouter"] != "sk-or-freekey" {
		t.Errorf("expected key sk-or-freekey in keys.yaml, got %s", keys["openrouter"])
	}
	expectedModel := models.DefaultModel("openrouter")
	if !strings.Contains(content, "model: "+expectedModel) {
		t.Errorf("expected default openrouter model %s, got:\n%s", expectedModel, content)
	}
}

func TestRunSetupFreeModeUseSavedKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)
	akp := ageKeyPath(dir)

	// Pre-seed the openrouter key directly
	if err := config.SaveKeys(kp, akp, map[string]string{"openrouter": "sk-or-saved"}); err != nil {
		t.Fatalf("failed to seed keys: %v", err)
	}

	// Free mode: press enter to reuse saved key
	var out bytes.Buffer
	err := Run(strings.NewReader("\n"), &out, configPath, kp, akp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keys, err2 := config.LoadKeys(kp, akp)
	if err2 != nil {
		t.Fatalf("LoadKeys failed: %v", err2)
	}
	if keys["openrouter"] != "sk-or-saved" {
		t.Errorf("expected saved openrouter key reused, got %s", keys["openrouter"])
	}
	// Prompt should show masked key
	if !strings.Contains(out.String(), "sk-o***") {
		t.Errorf("expected masked key in output, got:\n%s", out.String())
	}
}

func TestRunSetupFreeModeNoProviderPrompt(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)
	akp := ageKeyPath(dir)

	// Input has only an API key — no provider number
	input := strings.NewReader("sk-or-test\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, kp, akp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Provider:") {
		t.Error("free mode should not show provider prompt")
	}
	if strings.Contains(output, "Model [") {
		t.Error("free mode should not show model prompt")
	}
}
