package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Entry struct {
	Summary    string            `json:"summary"`
	FileHashes map[string]string `json:"file_hashes"`
}

type Cache struct {
	path    string
	entries map[string]*Entry
}

func New(path string) *Cache {
	c := &Cache{
		path:    path,
		entries: make(map[string]*Entry),
	}
	c.load()
	return c
}

func (c *Cache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &c.entries)
}

func (c *Cache) save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

func (c *Cache) Get(repoPath string) (*Entry, bool) {
	entry, ok := c.entries[repoPath]
	return entry, ok
}

func (c *Cache) Set(repoPath, summary string, hashes map[string]string) error {
	c.entries[repoPath] = &Entry{
		Summary:    summary,
		FileHashes: hashes,
	}
	return c.save()
}

func (c *Cache) NeedsRefresh(repoPath string, currentHashes map[string]string) bool {
	entry, ok := c.entries[repoPath]
	if !ok {
		return true
	}
	for file, hash := range currentHashes {
		if entry.FileHashes[file] != hash {
			return true
		}
	}
	for file := range entry.FileHashes {
		if _, ok := currentHashes[file]; !ok {
			return true
		}
	}
	return false
}

func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot hash file: %w", err)
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
