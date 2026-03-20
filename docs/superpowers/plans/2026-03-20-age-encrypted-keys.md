# Age-Encrypted Keys Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Encrypt API keys at rest using `filippo.io/age` so keys.yaml.age is an opaque ASCII-armored blob, with auto-generated keypair on first save.

**Architecture:** New `config/crypto.go` provides `ensureAgeKey`, `encryptKeys`, `decryptKeys`. Existing `SaveKeys`/`LoadKeys` in `config/config.go` gain an `ageKeyPath` param and call the crypto functions. All config paths move from `~/.config/commodo/` to `~/.commodo/`. Callers (`setup.go`, `main.go`) pass the new param.

**Tech Stack:** Go, `filippo.io/age`, existing `config`, `setup` packages.

---

## Chunk 1: crypto.go — age key management and encrypt/decrypt

### Task 1: Add age dependency and create crypto.go with tests

**Files:**
- Create: `config/crypto.go`
- Create: `config/crypto_test.go`

- [ ] **Step 1: Add `filippo.io/age` dependency**

```bash
go get filippo.io/age
```

- [ ] **Step 2: Write tests for `ensureAgeKey` in `config/crypto_test.go`**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureAgeKeyGeneratesNewKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "age-key.txt")

	identity, err := ensureAgeKey(keyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity == nil {
		t.Fatal("expected identity, got nil")
	}

	// File should exist with restricted permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("key file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}
}

func TestEnsureAgeKeyIdempotent(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "age-key.txt")

	id1, _ := ensureAgeKey(keyPath)
	id2, _ := ensureAgeKey(keyPath)

	if id1.Recipient().String() != id2.Recipient().String() {
		t.Error("expected same identity on second call, got different")
	}
}

func TestEnsureAgeKeyDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "age-key.txt")

	ensureAgeKey(keyPath)
	data1, _ := os.ReadFile(keyPath)

	ensureAgeKey(keyPath)
	data2, _ := os.ReadFile(keyPath)

	if string(data1) != string(data2) {
		t.Error("ensureAgeKey overwrote existing key file")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./config/ -run "TestEnsureAgeKey" -v
```
Expected: FAIL — `ensureAgeKey` undefined.

- [ ] **Step 4: Implement `ensureAgeKey` in `config/crypto.go`**

```go
package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// loadAgeKey reads and parses an existing age identity from keyPath.
// Returns an error if the file does not exist or cannot be parsed.
func loadAgeKey(keyPath string) (*age.X25519Identity, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read age key %s: %w", keyPath, err)
	}
	identities, err := age.ParseIdentities(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("cannot parse age key %s: %w", keyPath, err)
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities found in %s", keyPath)
	}
	id, ok := identities[0].(*age.X25519Identity)
	if !ok {
		return nil, fmt.Errorf("unexpected identity type in %s", keyPath)
	}
	return id, nil
}

// ensureAgeKey loads an existing age identity from keyPath, or generates a new
// X25519 keypair and writes it in standard age key format if the file does not exist.
// Idempotent: never overwrites an existing key file.
func ensureAgeKey(keyPath string) (*age.X25519Identity, error) {
	if id, err := loadAgeKey(keyPath); err == nil {
		return id, nil
	}

	// Only generate if the file does not exist
	if _, err := os.Stat(keyPath); err == nil {
		// File exists but failed to parse — don't overwrite
		return loadAgeKey(keyPath)
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("cannot generate age key: %w", err)
	}

	content := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
		time.Now().UTC().Format(time.RFC3339),
		identity.Recipient().String(),
		identity.String(),
	)

	if err := os.WriteFile(keyPath, []byte(content), 0600); err != nil {
		return nil, fmt.Errorf("cannot write age key: %w", err)
	}

	return identity, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./config/ -run "TestEnsureAgeKey" -v
```
Expected: all 3 PASS.

- [ ] **Step 6: Write tests for `encryptKeys` and `decryptKeys`**

Append to `config/crypto_test.go`:

```go
func TestEncryptDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "age-key.txt")

	identity, _ := ensureAgeKey(keyPath)
	plaintext := []byte("openai: sk-test123\nanthropic: sk-ant456\n")

	ciphertext, err := encryptKeys(plaintext, identity)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Ciphertext should be ASCII-armored (starts with age header)
	if !bytes.Contains(ciphertext, []byte("-----BEGIN AGE ENCRYPTED FILE-----")) {
		t.Error("expected ASCII-armored output")
	}

	decrypted, err := decryptKeys(ciphertext, identity)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	dir := t.TempDir()
	id1, _ := ensureAgeKey(filepath.Join(dir, "key1.txt"))
	id2, _ := ensureAgeKey(filepath.Join(dir, "key2.txt"))

	ciphertext, _ := encryptKeys([]byte("secret"), id1)
	_, err := decryptKeys(ciphertext, id2)
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	dir := t.TempDir()
	identity, _ := ensureAgeKey(filepath.Join(dir, "key.txt"))

	_, err := decryptKeys([]byte("not-encrypted-data"), identity)
	if err == nil {
		t.Error("expected error decrypting corrupted data")
	}
}
```

- [ ] **Step 7: Run tests to verify they fail**

```bash
go test ./config/ -run "TestEncrypt|TestDecrypt" -v
```
Expected: FAIL — `encryptKeys`, `decryptKeys` undefined.

- [ ] **Step 8: Implement `encryptKeys` and `decryptKeys` in `config/crypto.go`**

Add to `config/crypto.go`:

```go
// encryptKeys encrypts plaintext using the recipient derived from the identity.
// Returns ASCII-armored ciphertext.
func encryptKeys(plaintext []byte, identity *age.X25519Identity) ([]byte, error) {
	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)
	w, err := age.Encrypt(armorWriter, identity.Recipient())
	if err != nil {
		return nil, fmt.Errorf("cannot create age encryptor: %w", err)
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("cannot write encrypted data: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("cannot finalize encryption: %w", err)
	}
	if err := armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("cannot finalize armor: %w", err)
	}
	return buf.Bytes(), nil
}

// decryptKeys decrypts ASCII-armored ciphertext using the identity.
func decryptKeys(ciphertext []byte, identity *age.X25519Identity) ([]byte, error) {
	armorReader := armor.NewReader(bytes.NewReader(ciphertext))
	r, err := age.Decrypt(armorReader, identity)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt keys: %w", err)
	}
	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read decrypted data: %w", err)
	}
	return plaintext, nil
}
```

- [ ] **Step 9: Run all crypto tests**

```bash
go test ./config/ -run "TestEnsureAgeKey|TestEncrypt|TestDecrypt" -v
```
Expected: all 6 PASS.

- [ ] **Step 10: Commit**

```bash
git add config/crypto.go config/crypto_test.go go.mod go.sum
git commit -m "feat(config): add age encryption primitives for key storage"
```

---

## Chunk 2: Update config.go — encrypted SaveKeys/LoadKeys and path changes

### Task 2: Update SaveKeys, LoadKeys, Load signatures and config paths

**Files:**
- Modify: `config/config.go`
- Modify: `config/config_test.go`

- [ ] **Step 1: Write new tests for encrypted SaveKeys/LoadKeys in `config/config_test.go`**

Replace `TestLoadResolvesKeyFromKeysFile`, `TestLoadMissingAPIKey`, and `TestLoadFallsBackToInlineKey` with encrypted versions. Also add new tests. The full updated `config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/garciasdos/commodo/models"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: openai\nmodel: gpt-4o-mini\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.Provider)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", cfg.Model)
	}
}

func TestLoadDefaultModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: deepseek\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := models.DefaultModel("deepseek")
	if cfg.Model != expected {
		t.Errorf("expected default model %s, got %s", expected, cfg.Model)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := LoadFrom("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadMissingProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("model: gpt-4o-mini\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestLoadInvalidProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("provider: fakellm\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}

func TestSaveAndLoadKeysRoundTrip(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	original := map[string]string{"openai": "sk-test123", "anthropic": "sk-ant456"}
	if err := SaveKeys(keysPath, ageKeyPath, original); err != nil {
		t.Fatalf("SaveKeys failed: %v", err)
	}

	loaded, err := LoadKeys(keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("LoadKeys failed: %v", err)
	}

	for k, v := range original {
		if loaded[k] != v {
			t.Errorf("key %s: expected %s, got %s", k, v, loaded[k])
		}
	}
}

func TestLoadKeysMissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "nonexistent.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	keys, err := LoadKeys(keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty map, got %v", keys)
	}
}

func TestLoadKeysExistButAgeKeyMissing(t *testing.T) {
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	// Save keys (creates age-key.txt)
	SaveKeys(keysPath, ageKeyPath, map[string]string{"openai": "sk-test"})

	// Remove age key
	os.Remove(ageKeyPath)

	_, err := LoadKeys(keysPath, ageKeyPath)
	if err == nil {
		t.Fatal("expected error when age key is missing but keys file exists")
	}
}

func TestLoadResolvesKeyFromEncryptedKeysFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	os.WriteFile(configPath, []byte("provider: openai\nmodel: gpt-4o-mini\n"), 0644)
	SaveKeys(keysPath, ageKeyPath, map[string]string{"openai": "sk-from-keys"})

	cfg, err := Load(configPath, keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "sk-from-keys" {
		t.Errorf("expected api_key from encrypted keys, got %s", cfg.APIKey)
	}
}

func TestLoadMissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	os.WriteFile(configPath, []byte("provider: openai\n"), 0644)

	_, err := Load(configPath, keysPath, ageKeyPath)
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
}

func TestLoadFallsBackToInlineKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	keysPath := filepath.Join(dir, "keys.yaml.age")
	ageKeyPath := filepath.Join(dir, "age-key.txt")

	// Config with api_key inline — should still work
	os.WriteFile(configPath, []byte("provider: openai\napi_key: sk-inline\nmodel: gpt-4o-mini\n"), 0644)

	cfg, err := Load(configPath, keysPath, ageKeyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "sk-inline" {
		t.Errorf("expected inline api_key preserved, got %s", cfg.APIKey)
	}
}
```

- [ ] **Step 2: Run tests to verify new tests fail (old signatures)**

```bash
go test ./config/ -v
```
Expected: compilation errors — `SaveKeys`, `LoadKeys`, `Load` have wrong signatures.

- [ ] **Step 3: Update `config/config.go`**

Replace the full file with:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garciasdos/commodo/models"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".commodo", "config.yaml")
}

func KeysPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".commodo", "keys.yaml.age")
}

func AgeKeyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".commodo", "age-key.txt")
}

// LoadKeys reads and decrypts the per-provider keys file.
// Returns empty map + nil error if the file does not exist (no keys saved yet).
// Returns an error if the file exists but cannot be decrypted.
func LoadKeys(path, ageKeyPath string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, nil
	}

	identity, err := loadAgeKey(ageKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load age key for decryption: %w", err)
	}

	plaintext, err := decryptKeys(data, identity)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt keys file: %w", err)
	}

	keys := map[string]string{}
	if err := yaml.Unmarshal(plaintext, &keys); err != nil {
		return nil, fmt.Errorf("cannot parse decrypted keys: %w", err)
	}
	return keys, nil
}

// SaveKeys encrypts and writes the per-provider keys file.
func SaveKeys(path, ageKeyPath string, keys map[string]string) error {
	plaintext, err := yaml.Marshal(keys)
	if err != nil {
		return fmt.Errorf("cannot marshal keys: %w", err)
	}

	identity, err := ensureAgeKey(ageKeyPath)
	if err != nil {
		return fmt.Errorf("cannot load age key for encryption: %w", err)
	}

	ciphertext, err := encryptKeys(plaintext, identity)
	if err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0600)
}

// Load reads config.yaml and resolves the API key from the encrypted keys file.
func Load(configPath, keysPath, ageKeyPath string) (*Config, error) {
	cfg, err := LoadFrom(configPath)
	if err != nil {
		return nil, err
	}

	if cfg.APIKey == "" {
		keys, err := LoadKeys(keysPath, ageKeyPath)
		if err != nil {
			return nil, err
		}
		cfg.APIKey = keys[cfg.Provider]
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("config: api_key not found (run commodo setup)")
	}

	return cfg, nil
}

// LoadFrom reads a config file. API key may be empty if stored in keys file.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Provider == "" {
		return nil, fmt.Errorf("config: provider is required")
	}
	if !models.ValidProviders()[cfg.Provider] {
		return nil, fmt.Errorf("config: invalid provider %q (use: anthropic, deepseek, openai, openrouter)", cfg.Provider)
	}
	if cfg.Model == "" {
		cfg.Model = models.DefaultModel(cfg.Provider)
	}

	return &cfg, nil
}
```

- [ ] **Step 4: Run config tests**

```bash
go test ./config/ -v
```
Expected: all PASS (crypto tests + config tests).

- [ ] **Step 5: Commit**

```bash
git add config/config.go config/config_test.go
git commit -m "feat(config): encrypt keys with age, move paths to ~/.commodo"
```

---

## Chunk 3: Update setup.go and main.go callers

### Task 3: Update setup.Run and main.go to pass ageKeyPath

**Files:**
- Modify: `setup/setup.go`
- Modify: `setup/setup_test.go`
- Modify: `main.go`

- [ ] **Step 1: Update `setup/setup.go`**

Change the `Run` signature and update `LoadKeys`/`SaveKeys` calls:

In the signature, change:
```go
func Run(in io.Reader, out io.Writer, configPath, keysPath string, modelOnly, freeMode bool) error {
```
to:
```go
func Run(in io.Reader, out io.Writer, configPath, keysPath, ageKeyPath string, modelOnly, freeMode bool) error {
```

Change the `LoadKeys` call at line 28:
```go
keys := config.LoadKeys(keysPath)
```
to:
```go
keys, err := config.LoadKeys(keysPath, ageKeyPath)
if err != nil {
    return fmt.Errorf("cannot load keys: %w", err)
}
```

Change the `SaveKeys` call at line 106:
```go
config.SaveKeys(keysPath, keys)
```
to:
```go
config.SaveKeys(keysPath, ageKeyPath, keys)
```

Also properly handle the `SaveKeys` error — change lines 103-107:
```go
// Persist the key for this provider
keys[providerName] = apiKey
if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err == nil {
    config.SaveKeys(keysPath, keys)
}
```
to:
```go
// Persist the key for this provider
keys[providerName] = apiKey
if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err != nil {
    return fmt.Errorf("cannot create keys directory: %w", err)
}
if err := config.SaveKeys(keysPath, ageKeyPath, keys); err != nil {
    return fmt.Errorf("cannot save keys: %w", err)
}
```

- [ ] **Step 2: Update `setup/setup_test.go`**

Update the helper and all call sites. Changes needed:

Change the `keysPath` helper:
```go
func keysPath(dir string) string {
	return filepath.Join(dir, "keys.yaml")
}
```
to:
```go
func keysPath(dir string) string {
	return filepath.Join(dir, "keys.yaml.age")
}

func ageKeyPath(dir string) string {
	return filepath.Join(dir, "age-key.txt")
}
```

Every `Run(...)` call needs `ageKeyPath(dir)` or `ageKeyPath(kp_dir)` inserted after `keysPath(dir)` or `kp`. The pattern:

For calls using `keysPath(dir)` directly:
```go
// Before:
Run(input, &out, configPath, keysPath(dir), false, false)
// After:
Run(input, &out, configPath, keysPath(dir), ageKeyPath(dir), false, false)
```

For calls using `kp` variable:
```go
// Before:
Run(..., configPath, kp, false, false)
// After — also need akp variable:
akp := ageKeyPath(dir)
Run(..., configPath, kp, akp, false, false)
```

Every `config.LoadKeys(keysPath(dir))` or `config.LoadKeys(kp)` call becomes:
```go
// Before:
config.LoadKeys(keysPath(dir))
// After:
config.LoadKeys(keysPath(dir), ageKeyPath(dir))
```
And now returns `(map[string]string, error)` — handle the error:
```go
keys, err := config.LoadKeys(kp, akp)
if err != nil {
    t.Fatalf("LoadKeys failed: %v", err)
}
```

Every `config.SaveKeys(kp, ...)` becomes `config.SaveKeys(kp, akp, ...)`.

Update all call sites in these tests:
- `TestRunSetup` — uses `keysPath(dir)` directly, add `ageKeyPath(dir)`
- `TestRunSetupDeepSeek` — uses `keysPath(dir)` directly
- `TestRunSetupAnthropic` — uses `keysPath(dir)` directly
- `TestRunSetupInvalidProvider` — uses `keysPath(dir)` directly
- `TestRunSetupEmptyAPIKey` — uses `keysPath(dir)` directly, also has `config.LoadKeys` call
- `TestRunSetupOutputMessages` — uses `keysPath(dir)` directly
- `TestRunSetupPersistsKey` — uses `kp` variable, has `config.LoadKeys` call
- `TestRunSetupModelOnly` — uses `kp` variable
- `TestRunSetupFreeMode` — uses `kp` variable, has `config.LoadKeys` call
- `TestRunSetupFreeModeUseSavedKey` — uses `kp` variable, has `config.SaveKeys` and `config.LoadKeys` calls
- `TestRunSetupFreeModeNoProviderPrompt` — uses `kp` variable

- [ ] **Step 3: Update `main.go`**

Change the setup call at line 32:
```go
if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), modelOnly, freeMode); err != nil {
```
to:
```go
if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath(), config.KeysPath(), config.AgeKeyPath(), modelOnly, freeMode); err != nil {
```

Change the `Load` call at line 66:
```go
cfg, err := config.Load(config.DefaultPath(), config.KeysPath())
```
to:
```go
cfg, err := config.Load(config.DefaultPath(), config.KeysPath(), config.AgeKeyPath())
```

Change the cache path at line 98 (also fix `os.Getenv("HOME")` to `os.UserHomeDir()` for cross-platform consistency):
```go
cachePath := filepath.Join(os.Getenv("HOME"), ".config", "commodo", "cache.json")
```
to:
```go
home, _ := os.UserHomeDir()
cachePath := filepath.Join(home, ".commodo", "cache.json")
```

- [ ] **Step 4: Run all tests**

```bash
go test ./... -v -race
```
Expected: all PASS.

- [ ] **Step 5: Build and smoke test**

```bash
go build -o commodo .
echo "sk-or-smoketest" | ./commodo setup --free
cat ~/.commodo/config.yaml
cat ~/.commodo/keys.yaml.age
cat ~/.commodo/age-key.txt
```
Expected:
- `config.yaml` has `provider: openrouter` and `model: ...` (no api_key)
- `keys.yaml.age` starts with `-----BEGIN AGE ENCRYPTED FILE-----`
- `age-key.txt` has standard age key format with `AGE-SECRET-KEY-...`

- [ ] **Step 6: Commit**

```bash
git add setup/setup.go setup/setup_test.go main.go
git commit -m "feat(setup): wire age encryption through setup and main"
```

---

## Chunk 4: Update README and CLAUDE.md

### Task 4: Update documentation for new paths and encryption

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update README.md**

In the Setup section, update the config path description from:
```
This creates `~/.config/commodo/config.yaml` with your provider and model. API keys are saved per-provider in `~/.config/commodo/keys.yaml`
```
to:
```
This creates `~/.commodo/config.yaml` with your provider and model. API keys are encrypted with age and saved per-provider in `~/.commodo/keys.yaml.age`
```

Update the manual config example from:
```
`~/.config/commodo/config.yaml`:
```
to:
```
`~/.commodo/config.yaml`:
```

Remove the `keys.yaml` manual example section entirely (users can't manually create encrypted files).

- [ ] **Step 2: Update CLAUDE.md**

Change the config description:
```
| `config/` | YAML config loading (~/.config/commodo/config.yaml) and per-provider key store (~/.config/commodo/keys.yaml) |
```
to:
```
| `config/` | YAML config loading (~/.commodo/config.yaml), age-encrypted key store (~/.commodo/keys.yaml.age), and crypto helpers |
```

Update the setup wizard section to mention that API keys are encrypted.

- [ ] **Step 3: Run all tests one final time**

```bash
go test ./... -race
```
Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: update paths to ~/.commodo and document age encryption"
```
