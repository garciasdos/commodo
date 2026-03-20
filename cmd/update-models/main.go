package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/garciasdos/commodo/models"
	"gopkg.in/yaml.v3"
)

const openRouterURL = "https://openrouter.ai/api/v1/models"

type openRouterResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID      string              `json:"id"`
	Pricing openRouterPricing   `json:"pricing"`
}

type openRouterPricing struct {
	Prompt string `json:"prompt"`
}

func main() {
	yamlPath, err := findYAMLPath()
	if err != nil {
		log.Fatalf("finding models-pricing.yaml: %v", err)
	}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Fatalf("reading %s: %v", yamlPath, err)
	}

	cfg, err := models.Parse(data)
	if err != nil {
		log.Fatalf("parsing config: %v", err)
	}

	orModels, err := fetchOpenRouterModels()
	if err != nil {
		log.Fatalf("fetching OpenRouter models: %v", err)
	}

	changed := false
	for name, provider := range cfg.Providers {
		cheapest, price, err := findCheapest(provider, orModels)
		if err != nil {
			log.Printf("warning: %s: %v", name, err)
			continue
		}
		if cheapest != provider.Default {
			log.Printf("%s: updating default %s -> %s (%.10f $/token)", name, provider.Default, cheapest, price)
			provider.Default = cheapest
			cfg.Providers[name] = provider
			changed = true
		} else {
			log.Printf("%s: default %s is already cheapest (%.10f $/token)", name, provider.Default, price)
		}
	}

	if !changed {
		log.Println("no changes needed")
		return
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("marshaling YAML: %v", err)
	}

	if err := os.WriteFile(yamlPath, out, 0644); err != nil {
		log.Fatalf("writing %s: %v", yamlPath, err)
	}
	log.Printf("updated %s", yamlPath)
}

func findYAMLPath() (string, error) {
	// Walk up from the executable or current directory to find the repo root
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "models", "models-pricing.yaml")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("models-pricing.yaml not found from %s", dir)
}

func fetchOpenRouterModels() ([]openRouterModel, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(openRouterURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", openRouterURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", openRouterURL, resp.StatusCode)
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return result.Data, nil
}

func findCheapest(provider models.Provider, orModels []openRouterModel) (string, float64, error) {
	cheapestPrice := math.MaxFloat64
	cheapestModel := ""

	for _, m := range orModels {
		// OpenRouter IDs are prefixed: "openai/gpt-4o-mini"
		if !strings.HasPrefix(m.ID, provider.Prefix+"/") {
			continue
		}
		modelID := strings.TrimPrefix(m.ID, provider.Prefix+"/")

		if !matchesAny(modelID, provider.Patterns) {
			continue
		}

		price, err := strconv.ParseFloat(m.Pricing.Prompt, 64)
		if err != nil || price <= 0 {
			continue
		}

		if price < cheapestPrice {
			cheapestPrice = price
			cheapestModel = modelID
		}
	}

	if cheapestModel == "" {
		return "", 0, fmt.Errorf("no matching models found")
	}
	return cheapestModel, cheapestPrice, nil
}

func matchesAny(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
