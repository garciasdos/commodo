package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/garciasdos/commodo/config"
	"github.com/garciasdos/commodo/models"
)

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}

// Run runs the interactive setup wizard.
// keysPath is the path to the persistent per-provider keys store.
// modelOnly skips provider and API key steps, updating only the model.
// freeMode skips provider and model prompts, configuring OpenRouter with its free default.
func Run(in io.Reader, out io.Writer, configPath, keysPath string, modelOnly, freeMode bool) error {
	scanner := bufio.NewScanner(in)
	keys := config.LoadKeys(keysPath)

	var providerName, apiKey, defaultModel string

	if freeMode {
		providerName = "openrouter"
		defaultModel = models.DefaultModel("openrouter")
	} else if modelOnly {
		// Load existing config to get current provider and key
		cfg, err := config.LoadFrom(configPath)
		if err != nil {
			return fmt.Errorf("cannot load existing config (run commodo setup first): %w", err)
		}
		providerName = cfg.Provider
		apiKey = cfg.APIKey
		defaultModel = models.DefaultModel(providerName)
		if cfg.Model != "" {
			defaultModel = cfg.Model
		}
	} else {
		providers := models.Providers()

		// Provider selection
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
	}

	// API key — runs for freeMode and normal path; skipped for modelOnly (key already loaded above)
	if !modelOnly {
		existingKey := keys[providerName]
		for {
			if existingKey != "" {
				fmt.Fprintf(out, "\n  API key for %s [%s]: ", providerName, maskKey(existingKey))
			} else {
				fmt.Fprintf(out, "\n  API key for %s: ", providerName)
			}
			if !scanner.Scan() {
				return fmt.Errorf("unexpected end of input")
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "" && existingKey != "" {
				apiKey = existingKey
				break
			}
			if input != "" {
				apiKey = input
				break
			}
			fmt.Fprintln(out, "  API key cannot be empty.")
		}

		// Persist the key for this provider
		keys[providerName] = apiKey
		if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err == nil {
			config.SaveKeys(keysPath, keys)
		}
	}

	// Model — skip prompt for freeMode
	model := defaultModel
	if !freeMode {
		fmt.Fprintf(out, "\n  Model [%s]: ", defaultModel)
		if scanner.Scan() {
			if m := strings.TrimSpace(scanner.Text()); m != "" {
				model = m
			}
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
