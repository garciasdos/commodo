package prompt

import (
	"strings"
	"testing"
)

func TestCommitMessagePrompt(t *testing.T) {
	sys, user := CommitMessage("diff content here", "A Go CLI tool")

	if !strings.Contains(sys, "conventional commit") {
		t.Error("system prompt should mention conventional commit")
	}
	if !strings.Contains(user, "diff content here") {
		t.Error("user prompt should contain the diff")
	}
	if !strings.Contains(user, "A Go CLI tool") {
		t.Error("user prompt should contain the project summary")
	}
}

func TestCommitMessagePromptNoSummary(t *testing.T) {
	sys, user := CommitMessage("diff content", "")

	if !strings.Contains(sys, "conventional commit") {
		t.Error("system prompt should mention conventional commit")
	}
	if !strings.Contains(user, "diff content") {
		t.Error("user prompt should contain the diff")
	}
	if strings.Contains(user, "Project context") {
		t.Error("user prompt should not contain project context section when summary is empty")
	}
}

func TestSummaryPrompt(t *testing.T) {
	sys, user := Summary("# My Project\nA cool tool")

	if sys == "" {
		t.Error("expected non-empty system prompt")
	}
	if !strings.Contains(user, "# My Project") {
		t.Error("user prompt should contain the file contents")
	}
}
