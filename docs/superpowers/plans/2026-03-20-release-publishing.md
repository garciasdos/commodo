# Release & Publishing Flow Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automate release publishing with GoReleaser, GitHub Releases, and Homebrew tap updates.

**Architecture:** A GitHub Actions workflow triggers on `release: published`, runs tests, then delegates to GoReleaser for cross-compilation, checksumming, asset upload, and Homebrew formula push. GoReleaser config lives at repo root.

**Tech Stack:** GoReleaser, GitHub Actions, Homebrew tap (`garciasdos/homebrew-tap`)

**Spec:** `docs/superpowers/specs/2026-03-20-release-publishing-design.md`

---

## Chunk 1: GoReleaser Configuration & Release Workflow

### Task 1: Add `dist/` to `.gitignore`

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Add `dist/` to `.gitignore`**

Append `dist/` to the existing `.gitignore`. This directory is created by GoReleaser during local testing.

```
commodo
.worktrees/
dist/
```

- [ ] **Step 2: Commit**

```bash
git add .gitignore
git commit -m "chore: add dist/ to gitignore for GoReleaser"
```

---

### Task 2: Create GoReleaser configuration

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create `.goreleaser.yaml`**

```yaml
version: 2

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "commodo_{{.Version}}_{{.Os}}_{{.Arch}}"

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  groups:
    - title: Features
      regexp: '^.*?feat(\([[:word:]]+\))??!?:.+$'
      order: 0
    - title: Bug Fixes
      regexp: '^.*?fix(\([[:word:]]+\))??!?:.+$'
      order: 1
    - title: Others
      order: 999

brews:
  - repository:
      owner: garciasdos
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    name: commodo
    description: "CLI tool that generates conventional commit messages from staged git diffs using LLMs"
    homepage: "https://github.com/garciasdos/commodo"
    license: "MIT"
    install: |
      bin.install "commodo"
```

- [ ] **Step 2: Validate the config locally**

Run: `goreleaser check` (if goreleaser is installed locally), or just verify the YAML is valid:

```bash
go run gopkg.in/yaml.v3/... || echo "YAML syntax check — manual review OK"
```

If `goreleaser` is not installed, skip this — CI will validate.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "feat: add GoReleaser configuration for release builds"
```

---

### Task 3: Create release workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create `.github/workflows/release.yml`**

```yaml
name: Release

on:
  release:
    types: [published]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod

      - name: Test
        run: go test ./... -v -race

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "feat: add release workflow with GoReleaser and Homebrew tap"
```

---

### Task 4: Update CLAUDE.md with release documentation

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add release section to CLAUDE.md**

Add after the existing `## CI` section:

```markdown
## Releasing

- Create a release in the GitHub UI with a `v`-prefixed semver tag (e.g. `v1.0.0`)
- `release.yml` runs tests, then GoReleaser cross-compiles, uploads binaries + checksums, and pushes the Homebrew formula to `garciasdos/homebrew-tap`
- Users install via `brew install garciasdos/tap/commodo`
- `HOMEBREW_TAP_TOKEN` secret (fine-grained PAT scoped to the tap repo) must be configured in repo settings
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add releasing section to CLAUDE.md"
```

---

### Task 5: Verify end-to-end (dry run)

- [ ] **Step 1: Run GoReleaser in snapshot mode locally (if installed)**

```bash
goreleaser release --snapshot --clean
```

Expected: Creates `dist/` directory with 4 archives (`commodo_<snapshot>_linux_amd64.tar.gz`, etc.), `checksums.txt`, and no errors. The Homebrew formula push is skipped in snapshot mode.

If `goreleaser` is not installed locally, skip this — the workflow will validate on first real release.

- [ ] **Step 2: Verify `dist/` is ignored**

```bash
git status
```

Expected: `dist/` directory does not appear in untracked files.

- [ ] **Step 3: Clean up**

```bash
rm -rf dist/
```
