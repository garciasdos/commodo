package output

import (
	"bytes"
	"testing"
)

func TestSuccess(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Success("feat(auth): add JWT token validation")
	got := buf.String()
	if got == "" {
		t.Fatal("expected output, got empty string")
	}
	// Check it contains the message
	if !contains(got, "feat(auth): add JWT token validation") {
		t.Errorf("expected message in output, got: %s", got)
	}
	// Check it contains the checkmark
	if !contains(got, "✓") {
		t.Errorf("expected ✓ in output, got: %s", got)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Error("No staged changes found.")
	got := buf.String()
	if !contains(got, "✗") {
		t.Errorf("expected ✗ in output, got: %s", got)
	}
	if !contains(got, "No staged changes found.") {
		t.Errorf("expected message in output, got: %s", got)
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Info("feat(auth): add JWT token validation")
	got := buf.String()
	if !contains(got, "ℹ") {
		t.Errorf("expected ℹ in output, got: %s", got)
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Warn("Project context updated")
	got := buf.String()
	if !contains(got, "⟳") {
		t.Errorf("expected ⟳ in output, got: %s", got)
	}
}

func TestSecondary(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Secondary("abc1234")
	got := buf.String()
	if !contains(got, "abc1234") {
		t.Errorf("expected hash in output, got: %s", got)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
