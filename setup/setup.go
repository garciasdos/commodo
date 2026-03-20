package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var providers = []struct {
	name         string
	defaultModel string
}{
	{"openai", "gpt-4o-mini"},
	{"deepseek", "deepseek-chat"},
	{"anthropic", "claude-sonnet-4-6-20250514"},
}

func Run(in io.Reader, out io.Writer, configPath string) error {
	scanner := bufio.NewScanner(in)

	// Provider selection
	var providerName, defaultModel string
	for {
		fmt.Fprintln(out, "\n  Provider:")
		for i, p := range providers {
			fmt.Fprintf(out, "    %d. %s\n", i+1, p.name)
		}
		fmt.Fprint(out, "\n  Choose [1-3]: ")

		if !scanner.Scan() {
			return fmt.Errorf("unexpected end of input")
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			providerName, defaultModel = providers[0].name, providers[0].defaultModel
		case "2":
			providerName, defaultModel = providers[1].name, providers[1].defaultModel
		case "3":
			providerName, defaultModel = providers[2].name, providers[2].defaultModel
		default:
			fmt.Fprintf(out, "  Invalid choice: %s\n", choice)
			continue
		}
		break
	}

	// API key
	var apiKey string
	for {
		fmt.Fprintf(out, "\n  API key for %s: ", providerName)
		if !scanner.Scan() {
			return fmt.Errorf("unexpected end of input")
		}
		apiKey = strings.TrimSpace(scanner.Text())
		if apiKey != "" {
			break
		}
		fmt.Fprintln(out, "  API key cannot be empty.")
	}

	// Model
	fmt.Fprintf(out, "\n  Model [%s]: ", defaultModel)
	model := defaultModel
	if scanner.Scan() {
		if m := strings.TrimSpace(scanner.Text()); m != "" {
			model = m
		}
	}

	// Write config
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	content := fmt.Sprintf("provider: %s\napi_key: %s\nmodel: %s\n", providerName, apiKey, model)
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	fmt.Fprintf(out, "\n  Config saved to %s\n", configPath)
	return nil
}
