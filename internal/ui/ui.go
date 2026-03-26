package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tiw302/ai-commit/internal/config"
)

// UI handles terminal styling.
type UI struct {
	Success string
	Error   string
	Warning string
	Info    string
	Prompt  string
}

// NewUI initializes UI with default colors.
func NewUI() *UI {
	return &UI{
		Success: "\033[32m",
		Error:   "\033[31m",
		Warning: "\033[33m",
		Info:    "\033[34m",
		Prompt:  "\033[36m",
	}
}
// PrintSuccess displays a success message.
func (u *UI) PrintSuccess(msg string) {
	fmt.Printf("%s✔ %s\033[0m\n", u.Success, msg)
}

// PrintError displays an error message.
func (u *UI) PrintError(msg string) {
	fmt.Printf("%s✖ Error: %s\033[0m\n", u.Error, msg)
}

// PrintInfo displays an info message.
func (u *UI) PrintInfo(msg string) {
	fmt.Printf("%sℹ %s\033[0m\n", u.Info, msg)
}

// ApplyConfig updates colors from configuration.
func (u *UI) ApplyConfig(cfgColors config.UIColors) {
	if cfgColors.Success != "" {
		u.Success = cfgColors.Success
	}
	if cfgColors.Error != "" {
		u.Error = cfgColors.Error
	}
	if cfgColors.Warning != "" {
		u.Warning = cfgColors.Warning
	}
	if cfgColors.Info != "" {
		u.Info = cfgColors.Info
	}
}

// LoadingSpinner animates a loading state in the terminal.
func LoadingSpinner(stopChan chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-stopChan:
			fmt.Print("\r") // Clear the line when the task is done.
			return
		default:
			fmt.Printf("\r\033[36m%s\033[0m AI is reading your code and generating a commit message...", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// PromptUser displays a prompt and reads input.
func (u *UI) PromptUser(label string, defaultValue string) string {
	var promptMsg string
	if defaultValue != "" {
		promptMsg = fmt.Sprintf("%s%s [%s]: \033[0m", u.Prompt, label, defaultValue)
	} else {
		promptMsg = fmt.Sprintf("%s%s: \033[0m", u.Prompt, label)
	}
	fmt.Print(promptMsg)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// AskForConfirmation prompts for the next action.
func (u *UI) AskForConfirmation() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s? Accept this commit? [y]es / [n]o / [e]dit / [r]egenerate: \033[0m", u.Prompt)
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input))
}
// OpenInEditor opens the message in the system's default editor.
func OpenInEditor(initialMessage string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ai-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(initialMessage); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Check for $EDITOR environment variable, fallback to 'vim' or 'nano'.
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to open editor (%s): %w", editor, err)
	}

	// Read the content after user editing.
	updatedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read updated commit message: %w", err)
	}

	return strings.TrimSpace(string(updatedContent)), nil
}
