package config

import (
	"bytes"
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
