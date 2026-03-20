# Age-Encrypted Keys Design

## Goal

Encrypt API keys at rest using `filippo.io/age` so that `~/.commodo/keys.yaml.age` is an opaque binary blob rather than plaintext YAML. The age private key is co-located at `~/.commodo/age-key.txt`. This protects against accidental exposure (git commits, backups, screenshots, terminal shares) — not full filesystem compromise.

Zero extra steps for the user: keypair is auto-generated on first save.

## File Layout

```
~/.commodo/
├── config.yaml          # provider + model (plaintext, unchanged)
├── keys.yaml.age        # age-encrypted per-provider API keys
└── age-key.txt          # age X25519 private key (0600)
```

All config paths move from `~/.config/commodo/` to `~/.commodo/`.

## Config Path Changes

- `config.DefaultPath()` → `~/.commodo/config.yaml`
- `config.KeysPath()` → `~/.commodo/keys.yaml.age`
- `config.AgeKeyPath()` (new) → `~/.commodo/age-key.txt`

## Encryption Flow

### SaveKeys

1. Marshal `map[string]string` to YAML bytes
2. If `age-key.txt` doesn't exist, generate X25519 keypair and write private key (0600)
3. Read public key (recipient) from the identity file
4. Encrypt YAML bytes with `age.Encrypt(recipient)`
5. Write ciphertext to `keys.yaml.age` (0600)

### LoadKeys

1. Read `keys.yaml.age`
2. Read `age-key.txt`, parse X25519 identity
3. Decrypt with `age.Decrypt(identity)`
4. Unmarshal YAML to `map[string]string`
5. If file doesn't exist, return empty map (same as current behavior)

## New File: `config/crypto.go`

Functions:

- `ensureAgeKey(ageKeyPath string) (*age.X25519Identity, error)` — generates keypair if missing, returns identity
- `encryptKeys(plaintext []byte, ageKeyPath string) ([]byte, error)` — encrypts bytes using the public key derived from the identity
- `decryptKeys(ciphertext []byte, ageKeyPath string) ([]byte, error)` — decrypts bytes using the identity

## Signature Changes

- `SaveKeys(path, ageKeyPath string, keys map[string]string) error` — gains `ageKeyPath` param
- `LoadKeys(path, ageKeyPath string) map[string]string` — gains `ageKeyPath` param

All callers (`setup.go`, `main.go`, tests) updated to pass `ageKeyPath`.

## Testing

- `config/crypto_test.go` — unit tests for `ensureAgeKey`, `encryptKeys`/`decryptKeys` round-trip, missing key file behavior
- Existing `config/config_test.go` — updated `Load`/`LoadKeys`/`SaveKeys` calls with temp age key path
- Existing `setup/setup_test.go` — updated `Run` calls; tests use `t.TempDir()` for all three paths

## Dependencies

- Add `filippo.io/age` to `go.mod`

## What Doesn't Change

- `config.yaml` format and loading (plaintext, no encryption)
- `setup.Run()` signature and behavior (just passes `ageKeyPath` to `SaveKeys`/`LoadKeys`)
- User experience: no new prompts, no passphrase, fully transparent
