package git

import (
	"errors"
	"testing"
)

// Mock executor for testing
type mockExecutor struct {
	outputs map[string]string
	errors  map[string]error
}

func (m *mockExecutor) Run(args ...string) (string, error) {
	key := args[0]
	for _, a := range args[1:] {
		key += " " + a
	}
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	return m.outputs[key], nil
}

func TestStagedDiff(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"diff --cached": "diff --git a/main.go b/main.go\n+fmt.Println(\"hello\")\n",
		},
	}
	g := New(mock)

	diff, err := g.StagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
}

func TestStagedDiffEmpty(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"diff --cached": "",
		},
	}
	g := New(mock)

	_, err := g.StagedDiff()
	if err == nil {
		t.Fatal("expected error for empty diff")
	}
}

func TestCommit(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"commit -m feat(auth): add login": "",
			"rev-parse --short HEAD":          "abc1234",
		},
	}
	g := New(mock)

	hash, err := g.Commit("feat(auth): add login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "abc1234" {
		t.Errorf("expected abc1234, got %s", hash)
	}
}

func TestCommitFailure(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{},
		errors: map[string]error{
			"commit -m bad message": errors.New("git error"),
		},
	}
	g := New(mock)

	_, err := g.Commit("bad message")
	if err == nil {
		t.Fatal("expected error for failed commit")
	}
}

func TestRepoRoot(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"rev-parse --show-toplevel": "/Users/diego/Code/my-repo",
		},
	}
	g := New(mock)

	root, err := g.RepoRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != "/Users/diego/Code/my-repo" {
		t.Errorf("expected /Users/diego/Code/my-repo, got %s", root)
	}
}

func TestSetGlobalAlias(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{
			"config --global alias.commodo !commodo": "",
		},
	}
	g := New(mock)

	err := g.SetGlobalAlias("commodo", "commodo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetGlobalAliasFailure(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string]string{},
		errors: map[string]error{
			"config --global alias.commodo !commodo": errors.New("git error"),
		},
	}
	g := New(mock)

	err := g.SetGlobalAlias("commodo", "commodo")
	if err == nil {
		t.Fatal("expected error for failed alias creation")
	}
}
