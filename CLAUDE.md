# Commodo

CLI tool that generates conventional commit messages from staged git diffs using LLMs.

## Commands

```bash
go build -o commodo .          # Build
go test ./... -v -race         # Test (all packages)
go test ./models/              # Test single package
go vet ./...                   # Lint
```

## Architecture

| Package | Purpose |
|---------|---------|
| `main.go` | CLI entry point, subcommand routing |
| `cache/` | Hash-based project summary cache |
| `config/` | YAML config loading (~/.commodo/config.yaml), age-encrypted key store (~/.commodo/keys.yaml.age), and crypto helpers |
| `git/` | Git CLI wrapper (diff, commit, rev-parse) |
| `models/` | Embedded provider defaults from `models-pricing.yaml` |
| `output/` | Colored terminal output (ANSI) |
| `prompt/` | LLM prompt construction for commit messages and summaries |
| `provider/` | LLM API clients: OpenAI, Anthropic, DeepSeek, OpenRouter |
| `setup/` | Interactive configuration wizard |
| `cmd/update-models/` | Standalone tool to fetch cheapest models from OpenRouter |

## Key patterns

- **Interface-based design**: `Provider` interface for LLMs, `Executor` interface for git (enables mocking)
- **Composition**: DeepSeek and OpenRouter embed `*OpenAI` to reuse logic (same OpenAI-compatible API, different base URLs)
- **Embedded YAML**: `models-pricing.yaml` is embedded via `//go:embed` — defaults change automatically via the `update-models` workflow
- **Wrapped errors**: always use `fmt.Errorf("context: %w", err)`

## Setup wizard

- `commodo setup` — full interactive wizard (provider → API key → model)
- `commodo setup --model` — skips provider/key, updates model only
- `commodo setup --free` — skips provider/model, configures OpenRouter free tier
- API keys are age-encrypted and persisted per-provider in `keys.yaml.age`; shown masked as default on re-selection

## Testing

- Hand-rolled mocks (no assertion libraries)
- `t.TempDir()` for file-based tests
- Tests must **not** hardcode model names from `models-pricing.yaml` — use `models.DefaultModel()` dynamically since defaults are updated automatically by CI
- Setup tests pass `configPath`, `keysPath`, and `ageKeyPath` (all in `t.TempDir()`)

## CI

- **ci.yml**: vet + test + cross-platform build (linux/darwin × amd64/arm64)
- **update-models.yml**: weekly cron fetches cheapest models from OpenRouter and opens a PR

## Releasing

- Create a release in the GitHub UI with a `v`-prefixed semver tag (e.g. `v1.0.0`)
- `release.yml` runs tests, then GoReleaser cross-compiles, uploads binaries + checksums, and pushes the Homebrew formula to `garciasdos/homebrew-tap`
- Users install via `brew install garciasdos/tap/commodo`
- `HOMEBREW_TAP_TOKEN` secret (fine-grained PAT scoped to the tap repo) must be configured in repo settings
