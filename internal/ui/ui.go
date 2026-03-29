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

// TUI state
type UI struct {
	Success string
	Error   string
	Warning string
	Info    string
	Prompt  string
}

// init UI with colors
func NewUI() *UI {
	return &UI{
		Success: "\033[32m",
		Error:   "\033[31m",
		Warning: "\033[33m",
		Info:    "\033[34m",
		Prompt:  "\033[36m",
	}
}

// helper prints
func (u *UI) PrintSuccess(msg string) { fmt.Printf("%s✔ %s\033[0m\n", u.Success, msg) }
func (u *UI) PrintError(msg string)   { fmt.Printf("%s✖ error: %s\033[0m\n", u.Error, msg) }
func (u *UI) PrintInfo(msg string)    { fmt.Printf("%sℹ %s\033[0m\n", u.Info, msg) }

// sync config colors
func (u *UI) ApplyConfig(cfg config.UIColors) {
	if cfg.Success != "" { u.Success = cfg.Success }
	if cfg.Error != ""   { u.Error = cfg.Error }
	if cfg.Warning != "" { u.Warning = cfg.Warning }
	if cfg.Info != ""    { u.Info = cfg.Info }
}

// loading spinner
func LoadingSpinner(stopChan chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-stopChan:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\r\033[36m%s\033[0m generating...", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// read user input
func (u *UI) PromptUser(label string, defaultValue string) string {
	msg := fmt.Sprintf("%s%s: \033[0m", u.Prompt, label)
	if defaultValue != "" {
		msg = fmt.Sprintf("%s%s [%s]: \033[0m", u.Prompt, label, defaultValue)
	}
	fmt.Print(msg)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" { return defaultValue }
	return input
}

// ask for confirmation
func (u *UI) AskForConfirmation() string {
	fmt.Printf("\n%s? commit? [y]es / [n]o / [e]dit / [r]egenerate: \033[0m", u.Prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input))
}

// open editor for message
func OpenInEditor(msg string) (string, error) {
	tmp, _ := os.CreateTemp("", "ai-commit-*.txt")
	defer os.Remove(tmp.Name())

	tmp.WriteString(msg)
	tmp.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" { editor = "vim" }

	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Run()

	updated, _ := os.ReadFile(tmp.Name())
	return strings.TrimSpace(string(updated)), nil
}
