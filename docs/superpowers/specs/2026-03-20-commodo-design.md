# Commodo — CLI Commit Message Generator

## Overview

Commodo is a Go CLI tool that generates conventional commit messages by sending staged git diffs and project context to an LLM. It auto-commits with the generated message.

**Binary name**: `commodo`
Users can symlink to `git-commodo` if they want `git commodo` as a subcommand.

## CLI Interface

### Usage

```
commodo            # generate message from staged changes, commit
commodo --dry-run  # print the generated message without committing
```

### Flags

- `--dry-run` / `-d` — print message, don't commit
- `--version` / `-v` — print version

### Execution Flow

1. Check there are staged changes (`git diff --cached`), exit with error if none
2. Compute SHA-256 hashes of README.md / CLAUDE.md in the repo root
3. Compare hashes against cache — if miss or mismatch, read files and call LLM to generate a project summary, then update cache
4. Send staged diff + cached project summary to LLM
5. Receive conventional commit message
6. Run `git commit -m "<message>"`
7. Print commit hash and message with color

## Config

**Location**: `~/.config/commodo/config.yaml`

```yaml
provider: deepseek        # deepseek | openai | anthropic
api_key: sk-...
model: deepseek-chat      # optional, provider-specific default
```

Required fields: `provider`, `api_key`. The `model` field falls back to a sensible default per provider.

## Project Cache

**Location**: `~/.config/commodo/cache.json`

```json
{
  "/Users/diego/Code/my-repo": {
    "summary": "A Go CLI tool that generates commit messages...",
    "file_hashes": {
      "README.md": "a1b2c3...",
      "CLAUDE.md": "d4e5f6..."
    }
  }
}
```

### Cache Logic

1. On each run, compute SHA-256 of README.md and CLAUDE.md in the repo root (skip if file doesn't exist)
2. Compare against stored hashes for that repo path
3. If match: use cached summary
4. If mismatch or no entry: read files, call `GenerateSummary`, store result + new hashes

Auto-invalidation only — no manual refresh flag needed.

## LLM Integration

### Provider Interface

```go
type Provider interface {
    GenerateSummary(context string) (string, error)
    GenerateCommitMessage(diff string, projectSummary string) (string, error)
}
```

Two distinct LLM calls:

1. **`GenerateSummary`** — called on first run or cache invalidation. Input: README + CLAUDE.md contents. Output: concise project summary (2-3 sentences).
2. **`GenerateCommitMessage`** — called every run. Input: `git diff --cached` + cached project summary. Output: conventional commit message.

### Supported Providers

All using direct HTTP REST API calls (no SDK dependencies):

- **DeepSeek** — default model: `deepseek-chat`
- **OpenAI** — default model: `gpt-4o-mini`
- **Anthropic** — default model: `claude-sonnet-4-6-20250514`

### Commit Message Format

The LLM is prompted to produce a single conventional commit subject line:

```
type(scope): description
```

- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`
- No body, no footer — subject line only
- Max ~72 characters

## Output & Colors

Standard terminal color conventions:

| Color  | Usage                                          |
|--------|------------------------------------------------|
| Red    | Errors (no staged changes, API failure, etc.)  |
| Green  | Success (commit created, hash + message)       |
| Yellow | Warnings (cache invalidated, regenerating)     |
| Cyan   | Info (dry-run output)                          |
| Gray   | Secondary info (commit hash, dim text)         |

### Example Outputs

```
# Success
✓ feat(auth): add JWT token validation
  abc1234

# Error
✗ No staged changes found. Stage files with git add first.

# Dry run
ℹ feat(auth): add JWT token validation
  (dry run — no commit created)

# Cache refresh
⟳ Project context updated for /Users/diego/Code/my-repo
```

## Error Handling

- No staged changes → red error, exit 1
- Missing config file → red error with setup instructions, exit 1
- Missing/invalid API key → red error, exit 1
- LLM API failure → red error with HTTP status, exit 1
- No README.md or CLAUDE.md → proceed without project context (no cache entry)

## Project Structure

```
commodo/
├── main.go                  # entry point, CLI flag parsing
├── go.mod
├── go.sum
├── config/
│   ├── config.go            # load/parse config.yaml
│   └── config_test.go
├── cache/
│   ├── cache.go             # project cache read/write/invalidation
│   └── cache_test.go
├── git/
│   ├── git.go               # git operations (staged diff, commit)
│   └── git_test.go
├── provider/
│   ├── provider.go          # Provider interface
│   ├── deepseek.go          # DeepSeek implementation
│   ├── openai.go            # OpenAI implementation
│   ├── anthropic.go         # Anthropic implementation
│   └── provider_test.go     # mocked HTTP tests for all providers
├── prompt/
│   ├── prompt.go            # system/user prompt construction
│   └── prompt_test.go
└── output/
    ├── output.go            # colored terminal output helpers
    └── output_test.go
```

## Testing Strategy

All tests use standard `go test` with no external frameworks.

- **config**: parse valid/invalid YAML, missing file handling
- **cache**: read/write/invalidation logic, hash comparison
- **git**: mock git commands via interface, test diff parsing and commit execution
- **provider**: HTTP mock server (`httptest`) to test request/response per provider
- **prompt**: verify prompt construction with various inputs
- **output**: verify correct ANSI color codes

## Dependencies

Minimal external dependencies:

- `gopkg.in/yaml.v3` — YAML config parsing
- Standard library for everything else (HTTP, JSON, crypto/sha256, os/exec)
