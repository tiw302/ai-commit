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
	return cmd.Run() == nil
}

// get staged diff
func GetStagedDiff(cfg *config.Config) (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repo")
	}

	var excludePatterns []string
	for _, pattern := range cfg.ExcludeFiles {
		excludePatterns = append(excludePatterns, ":!"+pattern)
	}

	args := append([]string{"diff", "--staged", "--"}, excludePatterns...)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
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
		return nil, fmt.Errorf("not a git repo")
	}

	cmd := exec.Command("git", "diff", "--name-only", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("not a git repo")
	}

	cmd := exec.Command("git", "ls-files", "--modified", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// stage file
func StageFile(file string) error {
	cmd := exec.Command("git", "add", file)
	return cmd.Run()
}

// unstage file
func UnstageFile(file string) error {
	cmd := exec.Command("git", "restore", "--staged", file)
	return cmd.Run()
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
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

// get recent commits for changelog
func GetRecentCommits(n int) (string, error) {
	if !IsRepo() {
		return "", fmt.Errorf("not a git repo")
	}

	cmd := exec.Command("git", "log", "-n", fmt.Sprintf("%d", n), "--pretty=format:- %s")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// install git hook
func InstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repo")
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
	return os.WriteFile(hookPath, []byte(hookContent), 0755)
}

// remove git hook
func UninstallHook() error {
	if !IsRepo() {
		return fmt.Errorf("not a git repo")
	}

	hookPath := ".git/hooks/prepare-commit-msg"
	return os.Remove(hookPath)
}
