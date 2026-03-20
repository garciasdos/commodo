# Setup --free Mode Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `commodo setup --free` that skips provider/model prompts and configures OpenRouter with its free default model, only asking for an API key.

**Architecture:** Add a `freeMode bool` parameter to `setup.Run()`. When true, the provider is hard-coded to `openrouter` and the model to its default (from `models.DefaultModel("openrouter")`); only the API key step runs. `main.go` detects `--free` in `os.Args` and passes the flag through. README documents the new flag.

**Tech Stack:** Go, existing `setup`, `models`, `config` packages.

---

## Chunk 1: setup.Run() signature and free-mode logic

### Task 1: Add freeMode param to setup.Run() and update all callers

**Files:**
- Modify: `setup/setup.go`
- Modify: `main.go` (caller)
- Modify: `setup/setup_test.go` (all existing calls need the new arg)

- [ ] **Step 1: Update `setup.Run` signature to accept `freeMode bool`**

In `setup/setup.go`, change the signature from:
```go
func Run(in io.Reader, out io.Writer, configPath, keysPath string, modelOnly bool) error {
```
to:
```go
func Run(in io.Reader, out io.Writer, configPath, keysPath string, modelOnly, freeMode bool) error {
```

- [ ] **Step 2: Restructure `setup.Run()` to add free-mode branch**

The current `else` block in `setup.go` contains both the provider selection loop **and** the API key prompt. To make the API key prompt run for `freeMode` too, extract the API key block out of the `else` branch and place it after the `if/else if/else` block, guarded by `!modelOnly`.

Replace the entire `if modelOnly { ... } else { ... }` block plus the API key persistence lines with the following structure:

```go
if freeMode {
    providerName = "openrouter"
    defaultModel = models.DefaultModel("openrouter")
} else if modelOnly {
    // Load existing config to get current provider and key
    cfg, err := config.LoadFrom(configPath)
    if err != nil {
        return fmt.Errorf("cannot load existing config (run commodo setup first): %w", err)
    }
    providerName = cfg.Provider
    apiKey = cfg.APIKey
    defaultModel = models.DefaultModel(providerName)
    if cfg.Model != "" {
        defaultModel = cfg.Model
    }
} else {
    providers := models.Providers()

    // Provider selection
    for {
        fmt.Fprintln(out, "\n  Provider:")
        for i, p := range providers {
            fmt.Fprintf(out, "    %d. %s\n", i+1, p.Name)
        }
        fmt.Fprintf(out, "\n  Choose [1-%d]: ", len(providers))

        if !scanner.Scan() {
            return fmt.Errorf("unexpected end of input")
        }
        choice := strings.TrimSpace(scanner.Text())

        idx := -1
        for i := range providers {
            if choice == fmt.Sprintf("%d", i+1) {
                idx = i
                break
            }
        }
        if idx < 0 {
            fmt.Fprintf(out, "  Invalid choice: %s\n", choice)
            continue
        }
        providerName = providers[idx].Name
        defaultModel = providers[idx].DefaultModel
        break
    }
}

// API key — runs for freeMode and normal path; skipped for modelOnly (key already loaded above)
if !modelOnly {
    existingKey := keys[providerName]
    for {
        if existingKey != "" {
            fmt.Fprintf(out, "\n  API key for %s [%s]: ", providerName, maskKey(existingKey))
        } else {
            fmt.Fprintf(out, "\n  API key for %s: ", providerName)
        }
        if !scanner.Scan() {
            return fmt.Errorf("unexpected end of input")
        }
        input := strings.TrimSpace(scanner.Text())
        if input == "" && existingKey != "" {
            apiKey = existingKey
            break
        }
        if input != "" {
            apiKey = input
            break
        }
        fmt.Fprintln(out, "  API key cannot be empty.")
    }

    // Persist the key for this provider
    keys[providerName] = apiKey
    if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err == nil {
        config.SaveKeys(keysPath, keys)
    }
}
```

After this block, the model prompt and config write remain. Wrap the model prompt to skip it for `freeMode`:

```go
model := defaultModel
if !freeMode {
    fmt.Fprintf(out, "\n  Model [%s]: ", defaultModel)
    if scanner.Scan() {
        if m := strings.TrimSpace(scanner.Text()); m != "" {
            model = m
        }
    }
}
```

- [ ] **Step 3: Fix all callers of `setup.Run()` in `main.go` and `setup/setup_test.go`**

In `main.go`, the call is:
```go
if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), modelOnly); err != nil {
```
Update to pass `false` for `freeMode` (will be updated to the real flag in Task 2):
```go
if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), modelOnly, false); err != nil {
```

In `setup/setup_test.go`, every call to `Run(...)` has 5 args — add `false` as the last arg to each:
```go
Run(input, &out, configPath, keysPath(dir), false, false)
```
Update all existing call sites (search for `Run(` in the file to find them all).

- [ ] **Step 4: Run tests to confirm existing tests still pass**

```bash
go test ./setup/ -v -race
```
Expected: all existing tests PASS, no compilation errors.

- [ ] **Step 5: Commit**

```bash
git add setup/setup.go setup/setup_test.go main.go
git commit -m "refactor(setup): add freeMode param to Run() (wired as false)"
```

---

### Task 2: Write and verify tests for --free mode, then wire flag in main.go

**Files:**
- Modify: `setup/setup_test.go`
- Modify: `main.go`

- [ ] **Step 1: Add `config` import to `setup/setup_test.go`**

The new `TestRunSetupFreeModeUseSavedKey` test uses `config.SaveKeys` to pre-seed a key without relying on provider ordering. Add the import:

```go
"github.com/garciasdos/commodo/config"
```

to the existing import block in `setup/setup_test.go`.

- [ ] **Step 2: Write the new tests in `setup/setup_test.go`**

Add at the end of the file:

```go
func TestRunSetupFreeMode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)

	// Free mode: only prompts for API key; provider and model are pre-set
	input := strings.NewReader("sk-or-freekey\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, kp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "provider: openrouter") {
		t.Errorf("expected provider openrouter, got:\n%s", content)
	}
	if !strings.Contains(content, "api_key: sk-or-freekey") {
		t.Errorf("expected api_key sk-or-freekey, got:\n%s", content)
	}
	expectedModel := models.DefaultModel("openrouter")
	if !strings.Contains(content, "model: "+expectedModel) {
		t.Errorf("expected default openrouter model %s, got:\n%s", expectedModel, content)
	}
}

func TestRunSetupFreeModeUseSavedKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)

	// Pre-seed the openrouter key directly — avoids relying on provider sort order
	if err := config.SaveKeys(kp, map[string]string{"openrouter": "sk-or-saved"}); err != nil {
		t.Fatalf("failed to seed keys: %v", err)
	}

	// Free mode: press enter to reuse saved key
	var out bytes.Buffer
	err := Run(strings.NewReader("\n"), &out, configPath, kp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if !strings.Contains(string(data), "api_key: sk-or-saved") {
		t.Errorf("expected saved openrouter key reused, got:\n%s", string(data))
	}
	// Prompt should show masked key
	if !strings.Contains(out.String(), "sk-o***") {
		t.Errorf("expected masked key in output, got:\n%s", out.String())
	}
}

func TestRunSetupFreeModeNoProviderPrompt(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	kp := keysPath(dir)

	// Input has only an API key — no provider number. If provider prompt appeared, scanner would consume the key as the choice and the test would fail or error.
	input := strings.NewReader("sk-or-test\n")
	var out bytes.Buffer

	err := Run(input, &out, configPath, kp, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "Provider:") {
		t.Error("free mode should not show provider prompt")
	}
	if strings.Contains(output, "Model [") {
		t.Error("free mode should not show model prompt")
	}
}
```

- [ ] **Step 3: Run the new tests to confirm they pass**

```bash
go test ./setup/ -run "TestRunSetupFreeMode" -v
```
Expected: all three new tests PASS (Task 1 already added the full free-mode logic).

- [ ] **Step 4: Wire `--free` flag detection in `main.go`**

In `main.go`, replace the current setup block:
```go
if len(os.Args) > 1 && os.Args[1] == "setup" {
    out := output.New(os.Stderr)
    modelOnly := len(os.Args) > 2 && os.Args[2] == "--model"
    if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), modelOnly, false); err != nil {
```
with:
```go
if len(os.Args) > 1 && os.Args[1] == "setup" {
    out := output.New(os.Stderr)
    var modelOnly, freeMode bool
    for _, arg := range os.Args[2:] {
        switch arg {
        case "--model":
            modelOnly = true
        case "--free":
            freeMode = true
        }
    }
    if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), modelOnly, freeMode); err != nil {
```

- [ ] **Step 5: Run all tests**

```bash
go test ./... -v -race
```
Expected: all tests PASS.

- [ ] **Step 6: Build and do a quick smoke test**

```bash
go build -o commodo .
echo "sk-or-smoketest" | ./commodo setup --free
```
Expected: prompts only for API key, then prints `Config saved to ...` with no provider or model prompts visible.

- [ ] **Step 7: Commit**

```bash
git add setup/setup.go setup/setup_test.go main.go
git commit -m "feat(setup): add --free mode for zero-prompt openrouter free tier setup"
```

---

## Chunk 2: README update

### Task 3: Document --free in README.md

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add `--free` to the Setup section of README.md**

In `README.md`, after the `commodo setup --model` block, add:

```markdown
To get started instantly with a free model (no cost, just an OpenRouter API key):

```bash
commodo setup --free
```

This skips provider and model selection and configures OpenRouter with its free default model (`nvidia/nemotron-3-super-120b-a12b:free`).
```

- [ ] **Step 2: Run tests one final time**

```bash
go test ./... -race
```
Expected: all PASS.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: document commodo setup --free in README"
```
