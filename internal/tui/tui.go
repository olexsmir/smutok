package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"olexsmir.xyz/smutok/internal/sync"
)

type Model struct {
	isQutting bool
	showErr   bool
	err       error

	sync sync.Strategy
}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg
		m.showErr = true
		return m, nil

	case tea.WindowSizeMsg:

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.isQutting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.isQutting {
		return ""
	}
	return "are you feeling smutok?"
}
