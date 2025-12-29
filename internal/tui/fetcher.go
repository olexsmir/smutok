package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"olexsmir.xyz/smutok/internal/store"
)

type fetchedArticles []store.Article

func (m *Model) fetchArticles(kind store.ArticleKind) tea.Cmd {
	return func() tea.Msg {
		articles, err := m.store.GetArticles(m.ctx, kind)
		if err != nil {
			return sendErr(err)
		}
		return fetchedArticles(articles)
	}
}

type triggerSync struct{}

func (m *Model) triggerSync() tea.Cmd {
	return func() tea.Msg { return m.startSync() }
}

type finishedSync struct{}

func (m *Model) startSync() tea.Cmd {
	return func() tea.Msg {
		return ""
	}
}

func (m *Model) setupTableWithArticles() {
	// clean up previous state
	m.table.SetRows([]table.Row{})

	columns := []table.Column{
		{Title: "date", Width: 10},
		{Title: "status", Width: 6},
		{Title: "author", Width: 14},
		{Title: "title", Width: m.table.Width() - 30},
	}

	rows := make([]table.Row, len(m.articles))
	for i, a := range m.articles {
		rows[i] = table.Row{
			m.toArticleDate(a.PublishedAt),
			m.toArticleStatus(a.IsRead, a.IsStarred),
			a.Author,
			a.Title,
		}
	}

	m.table.SetColumns(columns)
	m.table.SetRows(rows)
}

func (m *Model) toArticleStatus(isRead, isStarred bool) string {
	var out string
	if isRead {
		out += "✓ "
	}
	if isStarred {
		out += "★ "
	}
	return out
}

func (m *Model) toArticleDate(publishedAt int64) string {
	return time.Unix(publishedAt, 0).Format("2006/01/02")
}
