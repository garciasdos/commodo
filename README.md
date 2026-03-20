# commodo

Generate conventional commit messages from staged git diffs using an LLM.

## Install

```bash
go install github.com/garciasdos/commodo@latest
```

## Setup

Run the interactive setup to configure your LLM provider:

```bash
commodo setup
```

This creates `~/.config/commodo/config.yaml` with your provider, API key, and model. API keys are saved per-provider in `~/.config/commodo/keys.yaml` — when you switch providers and come back, your key is pre-filled.

To change only the model without re-entering your provider and API key:

```bash
commodo setup --model
```

To get started instantly with a free model (no cost, just an OpenRouter API key):

```bash
commodo setup --free
```

This skips provider and model selection and configures OpenRouter with its free default model.

You can also create the config manually:

```yaml
provider: openrouter
api_key: sk-or-your-key-here
model: nvidia/nemotron-3-super-120b-a12b:free
```

## Usage

Stage your changes and run commodo:

```bash
git add .
commodo
```

Preview the generated message without committing:

```bash
commodo --dry-run
```

Check the version:

```bash
commodo --version
```

## Providers

| Provider   | Default Model                             | Notes       |
|------------|-------------------------------------------|-------------|
| openrouter | nvidia/nemotron-3-super-120b-a12b:free    | Free tier   |
| openai     | gpt-5-nano                                |             |
| deepseek   | deepseek-chat-v3.1                        |             |
| anthropic  | claude-haiku-4-5-20251001                 |             |

Default models are updated automatically each week via the `update-models` workflow.

## How it works

1. Reads your staged `git diff`
2. Checks for `README.md` and `CLAUDE.md` in the repo root to build project context (cached with hash-based invalidation)
3. Sends the diff and context to your configured LLM
4. Commits with the generated conventional commit message

## Contributing

```bash
git clone https://github.com/garciasdos/commodo.git
cd commodo
go test ./...
go build -o commodo .
```

PRs welcome. Please include tests for new functionality.

## License

MIT
