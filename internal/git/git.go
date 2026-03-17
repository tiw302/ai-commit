package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsRepo checks if the current directory is a git repository.
func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// GetStagedDiff retrieves the diff of staged files.
func GetStagedDiff() (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	excludePatterns := []string{
		":!package-lock.json",
		":!go.sum",
		":!*.svg",
		":!*.png",
		":!*.jpg",
		":!*.pdf",
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
