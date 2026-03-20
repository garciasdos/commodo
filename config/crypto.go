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
