package tui

import (
	"context"

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
	ctx       context.Context
	state     AppState
	isQutting bool
	isSyncing bool
	showErr   bool
	err       error

	syncer Syncer
	store  *store.Sqlite

	articlesKind store.ArticleKind
	articles     []store.Article
	contents     string

	glamur   *glamour.TermRenderer
	table    table.Model    // feeds, articles feed
	viewport viewport.Model // article content reader
}

func NewModel(
	ctx context.Context,
	syncer Syncer,
	sqlite *store.Sqlite,
) *Model {
	var err error
	tbl := table.New(table.WithFocused(true))
	vp := viewport.New(0, 0)
	g, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle())

	return &Model{
		ctx:          ctx,
		err:          err,
		syncer:       syncer,
		store:        sqlite,
		glamur:       g,
		table:        tbl,
		viewport:     vp,
		state:        ArticlesView,
		articlesKind: store.ArticleUnread,
		isQutting:    false,
		showErr:      false,
		articles:     nil,
		isSyncing:    false,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("smutok"),
		m.fetchArticles(m.articlesKind),
		m.triggerSync(),
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
		m.table.SetWidth(msg.Width - 2)
		m.viewport.Height = msg.Height
		m.viewport.Width = msg.Width
		return m, nil

	case fetchedArticles:
		m.articles = msg
		m.setupTableWithArticles()
		return m, nil

	case triggerSync:
		m.isSyncing = true
		return m, m.triggerSync()

	case finishedSync:
		m.isSyncing = false
		return m, m.fetchArticles(m.articlesKind)

	case tea.KeyMsg:
		// page specific keys
		switch m.state {
		case ArticlesView, FeedsView:
			m.table, cmd = m.table.Update(msg)
			switch msg.String() {
			case "enter":
				articleID := m.articles[m.table.Cursor()].ID
				article, err := m.store.GetArticleByID(m.ctx, articleID)
				if err != nil {
					return m, sendErr(err)
				}

				content, err := htmlToMarkdown.ConvertString(article.Content)
				if err != nil {
					return m, sendErr(err)
				}

				cont, err := m.glamur.Render(content)
				if err != nil {
					return m, sendErr(err)
				}

				m.state = ReadingView
				m.viewport.SetContent(cont)
				return m, nil
			}

		case ReadingView:
			m.viewport, cmd = m.viewport.Update(msg)
		}

		// global keys
		switch msg.String() {
		case "q":
			m.isQutting = true
			return m, tea.Quit
		case "s":
			return m, m.triggerSync()
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
