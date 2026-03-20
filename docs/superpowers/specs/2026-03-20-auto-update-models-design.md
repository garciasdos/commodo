# Auto-Update Default Models via GitHub Action

## Summary

A weekly GitHub Action that queries OpenRouter's free API for available models across all supported providers (OpenAI, Anthropic, DeepSeek), filters by name patterns (mini, haiku, chat), and creates a PR updating the default model per provider to the cheapest one by input token price.

## Single Source of Truth: `models-pricing.yaml`

A root-level `models-pricing.yaml` file replaces hardcoded default models in `config/config.go` and `setup/setup.go`. It defines:

- **patterns**: glob patterns to filter eligible models per provider (e.g., `gpt-*-mini*`, `claude-*-haiku-*`, `deepseek-chat*`)
- **default**: the currently selected cheapest model per provider

Go code embeds this file via `//go:embed` in a new `models` package. Both `config` and `setup` packages derive their provider lists and defaults from it.

## Refactoring

### New `models` package

- `models/models-pricing.yaml` — the YAML file (embedded)
- `models/models.go` — parses embedded YAML, exposes `DefaultModel(provider)`, `ValidProviders()`, and `Providers()`

### Changes to existing packages

- `config/config.go` — remove `defaultModels` and `validProviders` maps, use `models.DefaultModel()` and `models.ValidProviders()`
- `setup/setup.go` — remove hardcoded `providers` slice, use `models.Providers()` to build the interactive menu

## `cmd/update-models` Go Script

A standalone Go CLI at `cmd/update-models/main.go`:

1. Reads `models-pricing.yaml` from disk (not embedded — needs to write back)
2. Fetches `GET https://openrouter.ai/api/v1/models` (free, no auth)
3. For each provider, filters OpenRouter results by matching model IDs against `patterns` using glob matching
4. Picks the cheapest model by `pricing.prompt` (input token cost)
5. Updates `default` in `models-pricing.yaml` if a cheaper model is found
6. Writes the updated file back to disk

### Edge cases

- If OpenRouter is unreachable, exit with error (GHA will retry next week)
- If no models match patterns for a provider, keep current default
- Model IDs from OpenRouter are prefixed with provider namespace (e.g., `openai/gpt-4o-mini`) — strip the prefix before matching

## GitHub Actions Workflow

`.github/workflows/update-models.yml`:

- **Schedule**: weekly, Mondays at 9:00 UTC
- **Manual trigger**: `workflow_dispatch`
- **Steps**: checkout → setup-go → `go run ./cmd/update-models` → `peter-evans/create-pull-request@v7`
- **Branch**: `chore/update-default-models` (reused, so open PRs get updated)
- **No API keys needed**

## Testing

- `models/models_test.go` — test YAML parsing, `DefaultModel()`, `ValidProviders()`
- `cmd/update-models` — test OpenRouter response parsing and model selection logic with mock HTTP server
- Existing `config` and `setup` tests updated to work with the new `models` package
