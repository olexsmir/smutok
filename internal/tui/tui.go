package tui

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"olexsmir.xyz/smutok/internal/store"
)

type Syncer interface {
	Sync(ctx context.Context) error
}

type AppState int

const (
	ArticlesView AppState = iota
	FeedsView
	ReadingView
)

type Model struct {
	ctx context.Context

	state     AppState
	isQutting bool
	showErr   bool
	err       error

	syncer Syncer
	store  *store.Sqlite

	articles []store.Article

	glamur   *glamour.TermRenderer
	table    table.Model    // feeds, articles feed
	viewport viewport.Model // article content reader
}

func NewModel(
	ctx context.Context,
	syncer Syncer,
	store *store.Sqlite,
) *Model {
	tbl := table.New(table.WithFocused(true))
	vp := viewport.New(0, 0)

	return &Model{
		ctx:       ctx,
		syncer:    syncer,
		store:     store,
		glamur:    &glamour.TermRenderer{},
		table:     tbl,
		viewport:  vp,
		state:     ArticlesView,
		isQutting: false,
		showErr:   false,
		err:       nil,
		articles:  nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("smutok"),
		m.fetchArticles(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg
		m.showErr = true
		return m, nil

	case tea.WindowSizeMsg:
		m.table.SetHeight(msg.Height)
		m.table.SetWidth(msg.Width)
		m.viewport.Height = msg.Height
		m.viewport.Width = msg.Width
		return m, nil

	case fetchedArticles:
		m.articles = msg

		columns := []table.Column{
			// {Title: "read", Width: 4},
			// {Title: "stared", Width: 6},
			{Title: "author", Width: 14},
			{Title: "title", Width: m.table.Width() - 14},
		}

		rows := make([]table.Row, len(msg))
		for i, article := range msg {
			rows[i] = table.Row{article.Author, article.Title}
		}

		m.table.SetColumns(columns)
		m.table.SetRows(rows)

		slog.Debug("got articles")
		return m, nil

	case tea.KeyMsg:
		// page specific keys
		switch m.state {
		case ArticlesView, FeedsView:
			m.table, cmd = m.table.Update(msg)
		case ReadingView:
			m.viewport, cmd = m.viewport.Update(msg)
		}

		// global keys
		switch msg.String() {
		case "q":
			m.isQutting = true
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.isQutting {
		return ""
	}

	var out string
	switch m.state {
	case ArticlesView, FeedsView:
		out = m.table.View()
	case ReadingView:
		out = m.viewport.View()
	}
	return out
}
