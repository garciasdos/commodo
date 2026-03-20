# Age-Encrypted Keys Design

## Goal

Encrypt API keys at rest using `filippo.io/age` so that `~/.commodo/keys.yaml.age` is an ASCII-armored age-encrypted file rather than plaintext YAML. The age private key is co-located at `~/.commodo/age-key.txt`. This protects against accidental exposure (git commits, backups, screenshots, terminal shares) — not full filesystem compromise.

Zero extra steps for the user: keypair is auto-generated on first save.

## File Layout

```
~/.commodo/
├── config.yaml          # provider + model (plaintext, unchanged)
├── keys.yaml.age        # age-encrypted per-provider API keys (ASCII-armored)
└── age-key.txt          # age X25519 private key, standard age format (0600)
```

All config paths move from `~/.config/commodo/` to `~/.commodo/`.

## Config Path Changes

- `config.DefaultPath()` → `~/.commodo/config.yaml`
- `config.KeysPath()` → `~/.commodo/keys.yaml.age`
- `config.AgeKeyPath()` (new) → `~/.commodo/age-key.txt`

## age-key.txt Format

Standard age key file format for interoperability with the `age` CLI:

```
# created: 2026-03-20T12:00:00Z
# public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
AGE-SECRET-KEY-1QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ
```

Parsed back with `age.ParseIdentities()` (reads the standard format).

## Encryption Flow

### SaveKeys

1. Marshal `map[string]string` to YAML bytes
2. Call `ensureAgeKey(ageKeyPath)` — generates keypair if missing, returns `*age.X25519Identity`
3. Derive recipient from identity: `identity.Recipient()`
4. Encrypt YAML bytes with `age.Encrypt` using armored writer (`age.Armor`)
5. Write armored ciphertext to `keys.yaml.age` (0600)

### LoadKeys

1. Read `keys.yaml.age` — if file doesn't exist, return empty map + nil error
2. Read `age-key.txt`, parse identity with `age.ParseIdentities()`
3. Decrypt with `age.Decrypt(identity)` (handles armored input automatically)
4. Unmarshal YAML to `map[string]string`

### Error Handling

- `keys.yaml.age` missing → empty map, nil error (no keys saved yet)
- `keys.yaml.age` exists but `age-key.txt` missing → error (corrupted state)
- `keys.yaml.age` exists but decryption fails → error (corrupted or wrong key)
- Errors are surfaced to callers, never silently swallowed

## New File: `config/crypto.go`

Functions:

- `ensureAgeKey(ageKeyPath string) (*age.X25519Identity, error)` — if key file exists, parse and return identity; if not, generate new keypair, write in standard age format (0600), return identity. Idempotent: never overwrites existing key.
- `encryptKeys(plaintext []byte, identity *age.X25519Identity) ([]byte, error)` — encrypts bytes using `identity.Recipient()`, returns ASCII-armored ciphertext. Accepts identity directly (no file I/O).
- `decryptKeys(ciphertext []byte, identity *age.X25519Identity) ([]byte, error)` — decrypts armored ciphertext using identity. Accepts identity directly (no file I/O).

## Signature Changes

- `SaveKeys(path, ageKeyPath string, keys map[string]string) error` — gains `ageKeyPath` param
- `LoadKeys(path, ageKeyPath string) (map[string]string, error)` — gains `ageKeyPath` param, now returns error
- `Load(configPath, keysPath, ageKeyPath string) (*Config, error)` — gains `ageKeyPath` param
- `setup.Run(in, out, configPath, keysPath, ageKeyPath string, modelOnly, freeMode bool) error` — gains `ageKeyPath` param

All callers (`setup.go`, `main.go`, tests) updated accordingly.

## Testing

### `config/crypto_test.go`

- Round-trip: encrypt then decrypt returns identical plaintext
- `ensureAgeKey` generates key on first call, returns same identity on subsequent calls (idempotent)
- `ensureAgeKey` does not overwrite existing key file
- Decryption with wrong identity fails with error
- Corrupted / truncated ciphertext fails with error

### `config/config_test.go`

- `SaveKeys` + `LoadKeys` round-trip through encryption
- `LoadKeys` with missing `keys.yaml.age` returns empty map + nil error
- `LoadKeys` with existing `keys.yaml.age` but missing `age-key.txt` returns error
- `Load` resolves key from encrypted keys file
- All existing tests updated with temp age key path

### `setup/setup_test.go`

- All existing `Run` calls updated with temp `ageKeyPath`
- Free-mode and normal-mode tests verify keys end up encrypted (not plaintext)

## Dependencies

- Add `filippo.io/age` to `go.mod`

## What Doesn't Change

- `config.yaml` format and loading (plaintext, no encryption)
- User experience: no new prompts, no passphrase, fully transparent
