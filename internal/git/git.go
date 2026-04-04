package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tiw302/ai-commit/internal/config"
)

// check if git repo
func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		// If the command fails, it's likely not a git repo or git is not installed.
		// We don't want to return an error here, just indicate it's not a repo.
		return false
	}
	return true
}

// get staged diff
func GetStagedDiff(cfg *config.Config) (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repository")
	}

	var excludeArgs []string
	for _, pattern := range cfg.ExcludeFiles {
		excludeArgs = append(excludeArgs, ":!"+pattern)
	}

	// Construct the git command with exclude patterns
	args := append([]string{"diff", "--staged", "--"}, excludeArgs...)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		// If git diff returns an error (e.g., no staged changes, or other git error)
		// check for common cases like empty diff.
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 && strings.Contains(string(exitErr.Stderr), "no diff") {
				return "", fmt.Errorf("no staged changes to commit")
			}
		}
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := string(out)
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("no staged changes")
	}

	// truncate if too long
	runes := []rune(diff)
	if len(runes) > cfg.MaxDiffLength {
		diff = string(runes[:cfg.MaxDiffLength]) + "\n\n(diff truncated...)"
	}

	return diff, nil
}

// list staged files
func GetStagedFiles() ([]string, error) {
	if !IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--name-only", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list staged files: %w", err)
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// list unstaged files
func GetUnstagedFiles() ([]string, error) {
	if !IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "ls-files", "--modified", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list unstaged files: %w", err)
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// stage file
func StageFile(file string) error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	cmd := exec.Command("git", "add", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage file %s: %w", file, err)
	}
	return nil
}

// unstage file
func UnstageFile(file string) error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	cmd := exec.Command("git", "restore", "--staged", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unstage file %s: %w", file, err)
	}
	return nil
}

// detect commit scope
func DetectScope(files []string) string {
	scopes := make(map[string]int)

	for _, file := range files {
		file = strings.ReplaceAll(file, "\\", "/")

		if strings.HasSuffix(file, "_test.go") || strings.HasPrefix(file, "test/") {
			scopes["test"]++
		} else if strings.HasSuffix(file, ".md") {
			scopes["docs"]++
		} else if strings.HasSuffix(file, "go.mod") || strings.HasSuffix(file, "Makefile") {
			scopes["build"]++
		} else if strings.HasPrefix(file, ".github/") {
			scopes["ci"]++
		} else if strings.HasPrefix(file, "internal/ui/") {
			scopes["ui"]++
		} else if strings.HasPrefix(file, "internal/api/") {
			scopes["api"]++
		} else if strings.HasPrefix(file, "internal/config/") {
			scopes["config"]++
		} else if strings.HasPrefix(file, "cmd/") {
			scopes["cli"]++
		} else {
			parts := strings.Split(file, "/")
			if len(parts) > 1 && parts[0] != "internal" && parts[0] != "pkg" {
				scopes[parts[0]]++
			}
		}
	}

	if len(scopes) == 0 {
		return ""
	}

	var maxScore int
	var bestScope string
	for scope, score := range scopes {
		if score > maxScore {
			maxScore = score
			bestScope = scope
		}
	}

	return bestScope
}

// run git commit -m
func Commit(message string) error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	
	cmd := exec.Command("git", "commit", "-F", "-")
	cmd.Stdin = strings.NewReader(message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// get recent commits for changelog
func GetRecentCommits(n int) (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "log", "-n", fmt.Sprintf("%d", n), "--pretty=format:- %s")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get recent commits: %w", err)
	}

	return string(out), nil
}

// install git hook
func InstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	hookPath := ".git/hooks/prepare-commit-msg"
	hookContent := `#!/bin/sh
# ai-commit hook
if [ "$AI_COMMIT_SKIP" = "1" ]; then
    exit 0
fi
exec < /dev/tty
ai-commit --hook "$1" "$2" "$3"
`
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write git hook: %w", err)
	}
	return nil
}

// remove git hook
func UninstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	hookPath := ".git/hooks/prepare-commit-msg"
	if err := os.Remove(hookPath); err != nil {
		// If the file doesn't exist, it's not an error, just means it's already uninstalled.
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove git hook: %w", err)
		}
	}
	return nil
}
