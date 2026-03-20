package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMiss(t *testing.T) {
	dir := t.TempDir()
	c := New(filepath.Join(dir, "cache.json"))

	entry, hit := c.Get("/some/repo")
	if hit {
		t.Fatal("expected cache miss")
	}
	if entry != nil {
		t.Fatal("expected nil entry on miss")
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	c := New(path)

	hashes := map[string]string{"README.md": "abc123"}
	if err := c.Set("/some/repo", "A cool project", hashes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, hit := c.Get("/some/repo")
	if !hit {
		t.Fatal("expected cache hit")
	}
	if entry.Summary != "A cool project" {
		t.Errorf("expected summary 'A cool project', got %s", entry.Summary)
	}
	if entry.FileHashes["README.md"] != "abc123" {
		t.Errorf("expected hash abc123, got %s", entry.FileHashes["README.md"])
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	c1 := New(path)
	if err := c1.Set("/repo", "summary", map[string]string{"README.md": "xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Load from same file
	c2 := New(path)
	entry, hit := c2.Get("/repo")
	if !hit {
		t.Fatal("expected cache hit after reload")
	}
	if entry.Summary != "summary" {
		t.Errorf("expected 'summary', got %s", entry.Summary)
	}
}

func TestNeedsRefreshNoEntry(t *testing.T) {
	dir := t.TempDir()
	c := New(filepath.Join(dir, "cache.json"))

	if !c.NeedsRefresh("/repo", map[string]string{"README.md": "abc"}) {
		t.Fatal("expected refresh needed for missing entry")
	}
}

func TestNeedsRefreshHashMismatch(t *testing.T) {
	dir := t.TempDir()
	c := New(filepath.Join(dir, "cache.json"))
	c.Set("/repo", "old summary", map[string]string{"README.md": "old"}) //nolint

	if !c.NeedsRefresh("/repo", map[string]string{"README.md": "new"}) {
		t.Fatal("expected refresh needed for hash mismatch")
	}
}

func TestNeedsRefreshHashMatch(t *testing.T) {
	dir := t.TempDir()
	c := New(filepath.Join(dir, "cache.json"))
	c.Set("/repo", "summary", map[string]string{"README.md": "same"}) //nolint

	if c.NeedsRefresh("/repo", map[string]string{"README.md": "same"}) {
		t.Fatal("expected no refresh needed for matching hashes")
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	os.WriteFile(path, []byte("# Hello"), 0644)

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Same content = same hash
	hash2, _ := HashFile(path)
	if hash != hash2 {
		t.Error("expected same hash for same content")
	}
}

func TestHashFileMissing(t *testing.T) {
	_, err := HashFile("/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
