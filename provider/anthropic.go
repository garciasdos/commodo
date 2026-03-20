package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/garciasdos/commodo/prompt"
)

type Anthropic struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewAnthropic(apiKey, model, baseURL string) *Anthropic {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &Anthropic{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		client:  newHTTPClient(),
	}
}

func (a *Anthropic) GenerateSummary(context string) (string, error) {
	sys, user := prompt.Summary(context)
	return a.chat(sys, user)
}

func (a *Anthropic) GenerateCommitMessage(diff, projectSummary string) (string, error) {
	sys, user := prompt.CommitMessage(diff, projectSummary)
	return a.chat(sys, user)
}

func (a *Anthropic) chat(system, user string) (string, error) {
	body := map[string]any{
		"model":      a.model,
		"max_tokens": 256,
		"system":     system,
		"messages": []map[string]string{
			{"role": "user", "content": user},
		},
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", a.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}
	return result.Content[0].Text, nil
}
