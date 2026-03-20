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

This creates `~/.config/commodo/config.yaml` with your provider, API key, and model.

You can also create the config manually:

```yaml
provider: openai
api_key: sk-your-key-here
model: gpt-4o-mini
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

| Provider  | Default Model              |
|-----------|----------------------------|
| openai    | gpt-4o-mini                |
| deepseek  | deepseek-chat              |
| anthropic | claude-sonnet-4-6-20250514 |

## How it works

1. Reads your staged `git diff`
2. Checks for `README.md` and `CLAUDE.md` in the repo root to build project context (cached with hash-based invalidation)
3. Sends the diff and context to your configured LLM
4. Commits with the generated conventional commit message

## Contributing

```bash
git clone https://github.com/dgarcia/commodo.git
cd commodo
go test ./...
go build -o commodo .
```

PRs welcome. Please include tests for new functionality.

## License

MIT
