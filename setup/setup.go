package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/garciasdos/commodo/models"
)

func Run(in io.Reader, out io.Writer, configPath string) error {
	scanner := bufio.NewScanner(in)
	providers := models.Providers()

	// Provider selection
	var providerName, defaultModel string
	for {
		fmt.Fprintln(out, "\n  Provider:")
		for i, p := range providers {
			fmt.Fprintf(out, "    %d. %s\n", i+1, p.Name)
		}
		fmt.Fprintf(out, "\n  Choose [1-%d]: ", len(providers))

		if !scanner.Scan() {
			return fmt.Errorf("unexpected end of input")
		}
		choice := strings.TrimSpace(scanner.Text())

		idx := -1
		for i := range providers {
			if choice == fmt.Sprintf("%d", i+1) {
				idx = i
				break
			}
		}
		if idx < 0 {
			fmt.Fprintf(out, "  Invalid choice: %s\n", choice)
			continue
		}
		providerName = providers[idx].Name
		defaultModel = providers[idx].DefaultModel
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
