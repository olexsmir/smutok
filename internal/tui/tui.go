package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"olexsmir.xyz/smutok/internal/store"
)

type Syncer interface {
	Sync(ctx context.Context) error
}

type Model struct {
	ctx context.Context

	isQutting bool
	showErr   bool
	err       error

	syncer Syncer
	store  *store.Sqlite
}

func NewModel(
	ctx context.Context,
	syncer Syncer,
	store *store.Sqlite,
) *Model {
	return &Model{
		ctx:    ctx,
		syncer: syncer,
		store:  store,
	}
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
