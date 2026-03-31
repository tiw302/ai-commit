package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tiw302/ai-commit/internal/config"
)

// UI state
type UI struct {
	Success string
	Error   string
	Warning string
	Info    string
	Prompt  string
}

// styles
var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true)
	itemStyle   = lipgloss.NewStyle().PaddingLeft(2)
	selectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).PaddingLeft(0)
)

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
			fmt.Print("\r\033[K")
			return
		default:
			fmt.Printf("\r\033[36m%s\033[0m generating commit message...", spinner[i])
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

// confirm model
type confirmModel struct {
	choices  []string
	cursor   int
	selected string
}

func (m confirmModel) Init() tea.Cmd { return nil }
func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = "n"
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 { m.cursor-- }
		case "down", "j":
			if m.cursor < len(m.choices)-1 { m.cursor++ }
		case "enter":
			m.selected = string(m.choices[m.cursor][0])
			return m, tea.Quit
		case "y", "n", "e", "r":
			m.selected = msg.String()
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m confirmModel) View() string {
	s := titleStyle.Render("? commit message?") + "\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "❯"
			s += selectStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
		} else {
			s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
		}
	}
	s += "\n(press enter to select, or use shortcuts: y/n/e/r)\n"
	return s
}

// ask for confirmation with bubbletea
func (u *UI) AskForConfirmation() string {
	m := confirmModel{
		choices: []string{"yes", "no", "edit", "regenerate"},
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil { return "n" }
	return finalModel.(confirmModel).selected
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
