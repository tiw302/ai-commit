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
