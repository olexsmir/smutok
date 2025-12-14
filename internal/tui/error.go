package tui

import tea "github.com/charmbracelet/bubbletea"

type errMsg struct{ err error }

func (e errMsg) Error() string {
	return e.err.Error()
}

func sendErr(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}
