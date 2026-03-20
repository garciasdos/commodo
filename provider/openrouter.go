package provider

// OpenRouter uses the OpenAI-compatible API format

type OpenRouter struct {
	*OpenAI
}

func NewOpenRouter(apiKey, model, baseURL string) *OpenRouter {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenRouter{
		OpenAI: &OpenAI{
			apiKey:  apiKey,
			model:   model,
			baseURL: baseURL,
			client:  newHTTPClient(),
		},
	}
}
