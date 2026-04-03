package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileItem struct {
	Name     string
	IsStaged bool
	Selected bool
}

type stageModel struct {
	staged   []string
	unstaged []string
	items    []*FileItem
	cursor   int
	done     bool
}

func (m stageModel) Init() tea.Cmd { return nil }

func (m stageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 { m.cursor-- }
		case "down", "j":
			if m.cursor < len(m.items)-1 { m.cursor++ }
		case " ":
			m.items[m.cursor].Selected = !m.items[m.cursor].Selected
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m stageModel) View() string {
	s := titleStyle.Render("Select files to commit (space to toggle, enter to confirm)") + "\n\n"

	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = "❯"
		}

		checked := "[ ]"
		if item.Selected {
			checked = "[x]"
		}

		state := "unstaged"
		if item.IsStaged {
			state = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("staged")
		} else {
			state = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("unstaged")
		}

		line := fmt.Sprintf("%s %s %s (%s)", cursor, checked, item.Name, state)
		if m.cursor == i {
			s += selectStyle.Render(line) + "\n"
		} else {
			s += itemStyle.Render(line) + "\n"
		}
	}

	s += "\n(press q to cancel)\n"
	return s
}

func (u *UI) ShowStagingUI(staged, unstaged []string) []*FileItem {
	items := []*FileItem{}
	for _, f := range staged {
		items = append(items, &FileItem{Name: f, IsStaged: true, Selected: true})
	}
	for _, f := range unstaged {
		items = append(items, &FileItem{Name: f, IsStaged: false, Selected: false})
	}

	if len(items) == 0 {
		return nil
	}

	m := stageModel{
		staged:   staged,
		unstaged: unstaged,
		items:    items,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil { return nil }

	res := finalModel.(stageModel)
	if !res.done {
		return nil
	}

	return res.items
}
