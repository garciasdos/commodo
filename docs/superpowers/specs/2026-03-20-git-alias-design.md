# Git Alias Auto-Creation

## Summary

Automatically create a global git alias during `commodo setup` so users can run `git commodo` in any repo.

## Approach

Shell out to `git config --global alias.commodo '!commodo'` from Go using `exec.Command`. This leverages git's own tooling and matches the existing pattern in the `git` package.

The operation is idempotent — running setup twice overwrites the alias with the same value. No "already exists" check is needed.

## Changes

### `git` package — new method

Add `SetGlobalAlias(name, command string) error` as a method on `*Git`. It runs:

```
git config --global alias.<name> !<command>
```

Uses the existing `Executor` interface for testability. `exec.Command` passes args directly (no shell), so the `!` prefix requires no special quoting.

### `main.go` — call alias creation after setup

In `main.go`, after `setup.Run()` succeeds (around line 27), call `git.NewShell().SetGlobalAlias("commodo", "commodo")`.

- On success: print `Git alias created — you can now use "git commodo"` via `output.New(os.Stderr)`
- On failure: print a warning but do not fail setup. The alias is a convenience, not critical.

This keeps `setup.Run()` focused on config creation with no new dependencies. The `setup` package signature and tests remain unchanged.

### Tests

- Unit test for `SetGlobalAlias` using the existing `MockExecutor` pattern in `git/git_test.go`

## Non-goals

- Per-repo aliases (global only)
- User confirmation prompt (setup is already an opt-in action)
- Alias removal/cleanup on uninstall
