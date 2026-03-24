package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tiw302/ai-commit/internal/config"
)

// IsRepo checks if the current directory is a git repository.
func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// GetStagedDiff retrieves the diff of staged files.
func GetStagedDiff(cfg *config.Config) (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	var excludePatterns []string
	for _, pattern := range cfg.ExcludeFiles {
		excludePatterns = append(excludePatterns, ":!"+pattern)
	}

	args := append([]string{"diff", "--staged", "--"}, excludePatterns...)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := string(out)
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("no staged changes found; use 'git add' to stage files")
	}

	// Truncate diff if it's too long to save tokens and avoid API limits.
	runes := []rune(diff)
	if len(runes) > cfg.MaxDiffLength {
		diff = string(runes[:cfg.MaxDiffLength]) + "\n\n(diff truncated for length...)"
	}

	return diff, nil
}

// GetStagedFiles returns a list of files that are currently staged.
func GetStagedFiles() ([]string, error) {
	if !IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--name-only", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// DetectScope attempts to infer the Conventional Commit scope based on the file paths.
func DetectScope(files []string) string {
	scopes := make(map[string]int)

	for _, file := range files {
		// Normalize path separators
		file = strings.ReplaceAll(file, "\\", "/")

		if strings.HasPrefix(file, "internal/ui/") || strings.HasPrefix(file, "ui/") {
			scopes["ui"]++
		} else if strings.HasPrefix(file, "internal/api/") || strings.HasPrefix(file, "api/") {
			scopes["api"]++
		} else if strings.HasPrefix(file, "internal/config/") || strings.HasPrefix(file, "config/") {
			scopes["config"]++
		} else if strings.HasPrefix(file, "cmd/") {
			scopes["cli"]++
		} else if strings.HasSuffix(file, "_test.go") || strings.HasPrefix(file, "test/") {
			scopes["test"]++
		} else if strings.HasSuffix(file, ".md") {
			scopes["docs"]++
		} else if strings.HasSuffix(file, "go.mod") || strings.HasSuffix(file, "go.sum") || strings.HasSuffix(file, "Makefile") {
			scopes["build"]++
		} else if strings.HasPrefix(file, ".github/") {
			scopes["ci"]++
		} else {
			// Try to get the top-level directory as scope if it's not "internal" or "pkg"
			parts := strings.Split(file, "/")
			if len(parts) > 1 && parts[0] != "internal" && parts[0] != "pkg" && parts[0] != "src" {
				scopes[parts[0]]++
			}
		}
	}

	if len(scopes) == 0 {
		return ""
	}

	// Find the most frequent scope
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

// Commit executes the 'git commit -m' command with the provided commit message.
func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute git commit: %w", err)
	}
	return nil
}

// InstallHook sets up a git hook to run ai-commit automatically on commit.
func InstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	hookPath := ".git/hooks/prepare-commit-msg"
	hookContent := `#!/bin/sh
# ai-commit hook
# This hook was installed by ai-commit

# Check if we should skip the hook (e.g., if AI_COMMIT_SKIP=1)
if [ "$AI_COMMIT_SKIP" = "1" ]; then
    exit 0
fi

# Run ai-commit in hook mode
# Pass the commit message file path to ai-commit
exec < /dev/tty
ai-commit --hook "$1" "$2" "$3"
`

	err := os.WriteFile(hookPath, []byte(hookContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	return nil
}

// UninstallHook removes the git hook.
func UninstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	hookPath := ".git/hooks/prepare-commit-msg"
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook file: %w", err)
	}

	return nil
}
