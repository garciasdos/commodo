package provider

import (
	"fmt"
	"net/http"
	"time"
)

type Provider interface {
	GenerateSummary(context string) (string, error)
	GenerateCommitMessage(diff string, projectSummary string) (string, error)
}

func NewProvider(name, apiKey, model string) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAI(apiKey, model, ""), nil
	case "deepseek":
		return NewDeepSeek(apiKey, model, ""), nil
	case "anthropic":
		return NewAnthropic(apiKey, model, ""), nil
	case "openrouter":
		return NewOpenRouter(apiKey, model, ""), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 120 * time.Second}
}
