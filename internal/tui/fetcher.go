package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"olexsmir.xyz/smutok/internal/store"
)

type fetchedArticles []store.Article

func (m *Model) fetchArticles() tea.Cmd {
	return func() tea.Msg {
		articles, err := m.store.GetArticles(m.ctx)
		if err != nil {
			return sendErr(err)
		}
		return fetchedArticles(articles)
	}
}
