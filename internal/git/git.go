package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetStagedDiff retrieves the diff of staged files while filtering out 
// noise such as lock files, binaries, and images.
func GetStagedDiff() (string, error) {
	// Exclude non-text or high-volume files to minimize token usage.
	excludePatterns := []string{
		":!package-lock.json",
		":!go.sum",
		":!*.svg",
		":!*.png",
		":!*.jpg",
		":!*.pdf",
	}

	args := append([]string{"diff", "--staged", "--"}, excludePatterns...)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := string(out)
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("no staged changes found. please stage your changes with 'git add'")
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
