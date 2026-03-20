package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Executor interface {
	Run(args ...string) (string, error)
}

type ShellExecutor struct{}

func (s *ShellExecutor) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

type Git struct {
	exec Executor
}

func New(exec Executor) *Git {
	return &Git{exec: exec}
}

func NewShell() *Git {
	return &Git{exec: &ShellExecutor{}}
}

func (g *Git) StagedDiff() (string, error) {
	diff, err := g.exec.Run("diff", "--cached")
	if err != nil {
		return "", err
	}
	if diff == "" {
		return "", fmt.Errorf("no staged changes found")
	}
	return diff, nil
}

func (g *Git) Commit(message string) (string, error) {
	_, err := g.exec.Run("commit", "-m", message)
	if err != nil {
		return "", err
	}
	hash, err := g.exec.Run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (g *Git) RepoRoot() (string, error) {
	return g.exec.Run("rev-parse", "--show-toplevel")
}

func (g *Git) SetGlobalAlias(name, command string) error {
	_, err := g.exec.Run("config", "--global", "alias."+name, "!"+command)
	return err
}
