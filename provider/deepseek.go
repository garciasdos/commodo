package provider

// DeepSeek uses the OpenAI-compatible API format

type DeepSeek struct {
	*OpenAI
}

func NewDeepSeek(apiKey, model, baseURL string) *DeepSeek {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	return &DeepSeek{
		OpenAI: &OpenAI{
			apiKey:  apiKey,
			model:   model,
			baseURL: baseURL,
			client:  newHTTPClient(),
		},
	}
}
