package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/garciasdos/commodo/cache"
	"github.com/garciasdos/commodo/config"
	"github.com/garciasdos/commodo/git"
	"github.com/garciasdos/commodo/output"
	"github.com/garciasdos/commodo/provider"
	"github.com/garciasdos/commodo/setup"
)

var version = "dev"

func main() {
	// Handle subcommands before flag parsing
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		out := output.New(os.Stderr)
		if err := setup.Run(os.Stdin, os.Stderr, config.DefaultPath()); err != nil {
			out.Error(err.Error())
			os.Exit(1)
		}
		// Create git alias so "git commodo" works
		g := git.NewShell()
		execPath, err := os.Executable()
		if err != nil {
			execPath = "commodo" // fallback to bare name
		}
		if err := g.SetGlobalAlias("commodo", execPath); err != nil {
			out.Warn("Could not create git alias: " + err.Error())
			out.Secondary("You can still run commodo directly.")
		} else {
			out.Success("Git alias created — you can now use \"git commodo\"")
		}
		out.Success("Setup complete! Run commodo in a git repo with staged changes.")
		return
	}

	dryRun := flag.Bool("dry-run", false, "print generated message without committing")
	flag.BoolVar(dryRun, "d", false, "print generated message without committing")
	showVersion := flag.Bool("version", false, "print version")
	flag.BoolVar(showVersion, "v", false, "print version")
	flag.Parse()

	out := output.New(os.Stderr)

	if *showVersion {
		fmt.Printf("commodo %s\n", version)
		return
	}

	// Load config
	cfg, err := config.LoadFrom(config.DefaultPath())
	if err != nil {
		out.Error(err.Error())
		out.Secondary("Run: commodo setup")
		os.Exit(1)
	}

	// Initialize git
	g := git.NewShell()

	// Get staged diff
	diff, err := g.StagedDiff()
	if err != nil {
		out.Error("No staged changes found. Stage files with git add first.")
		os.Exit(1)
	}

	// Get repo root
	repoRoot, err := g.RepoRoot()
	if err != nil {
		out.Error("Not a git repository.")
		os.Exit(1)
	}

	// Initialize provider
	llm, err := provider.NewProvider(cfg.Provider, cfg.APIKey, cfg.Model)
	if err != nil {
		out.Error(err.Error())
		os.Exit(1)
	}

	// Initialize cache
	cachePath := filepath.Join(os.Getenv("HOME"), ".config", "commodo", "cache.json")
	c := cache.New(cachePath)

	// Compute file hashes for context files
	currentHashes := make(map[string]string)
	for _, name := range []string{"README.md", "CLAUDE.md"} {
		path := filepath.Join(repoRoot, name)
		if h, err := cache.HashFile(path); err == nil {
			currentHashes[name] = h
		}
	}

	// Get or refresh project summary
	var projectSummary string
	if len(currentHashes) > 0 {
		if c.NeedsRefresh(repoRoot, currentHashes) {
			// Read context files
			var contextContent string
			for _, name := range []string{"README.md", "CLAUDE.md"} {
				path := filepath.Join(repoRoot, name)
				data, err := os.ReadFile(path)
				if err == nil {
					contextContent += fmt.Sprintf("--- %s ---\n%s\n\n", name, string(data))
				}
			}

			out.Warn(fmt.Sprintf("Project context updated for %s", repoRoot))
			summary, err := llm.GenerateSummary(contextContent)
			if err != nil {
				out.Error(fmt.Sprintf("Failed to generate project summary: %v", err))
				os.Exit(1)
			}
			if err := c.Set(repoRoot, summary, currentHashes); err != nil {
				out.Warn(fmt.Sprintf("Failed to save cache: %v", err))
			}
			projectSummary = summary
		} else {
			entry, _ := c.Get(repoRoot)
			projectSummary = entry.Summary
		}
	}

	// Generate commit message
	message, err := llm.GenerateCommitMessage(diff, projectSummary)
	if err != nil {
		out.Error(fmt.Sprintf("Failed to generate commit message: %v", err))
		os.Exit(1)
	}

	// Dry run mode
	if *dryRun {
		out.Info(message)
		out.Secondary("(dry run — no commit created)")
		return
	}

	// Commit
	hash, err := g.Commit(message)
	if err != nil {
		out.Error(fmt.Sprintf("Commit failed: %v", err))
		os.Exit(1)
	}

	out.Success(message)
	out.Secondary(hash)
}
