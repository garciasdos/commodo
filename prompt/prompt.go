package prompt

import "fmt"

func CommitMessage(diff, projectSummary string) (system, user string) {
	system = `You are a commit message generator. Generate a single conventional commit message subject line.

Format: type(scope): description

Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf

Rules:
- Output ONLY the commit message, nothing else
- No body, no footer — subject line only
- Max 72 characters
- Use imperative mood ("add" not "added")
- Scope is optional — omit if the change spans multiple areas
- Be specific and concise`

	if projectSummary != "" {
		user = fmt.Sprintf("Project context:\n%s\n\nDiff:\n%s", projectSummary, diff)
	} else {
		user = fmt.Sprintf("Diff:\n%s", diff)
	}
	return system, user
}

func Summary(fileContents string) (system, user string) {
	system = `You are a project analyzer. Read the provided project files and produce a concise summary of what the project is about (2-3 sentences). Focus on purpose, tech stack, and domain. Output ONLY the summary, nothing else.`
	user = fileContents
	return system, user
}
