package store

import (
	"context"
	"fmt"
	"strings"
)

func (s *Sqlite) UpsertArticle(
	ctx context.Context,
	timestampUsec, feedID, title, content, author, href string,
	publishedAt int,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx,
		`insert or ignore into articles (id, feed_id, title, content, author, href, published_at) values (?, ?, ?, ?, ?, ?, ?)`,
		timestampUsec, feedID, title, content, author, href, publishedAt); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `insert or ignore into article_statuses (article_id) values (?)`, timestampUsec); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Sqlite) SyncReadStatus(ctx context.Context, ids []string) error {
	placeholders, args := buildPlaceholdersAndArgs(ids)
	query := fmt.Sprintf(`--sql
	update article_statuses
	set is_read = case when article_id in (%s)
		then false
		else true
	end`, placeholders)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Sqlite) SyncStarredStatus(ctx context.Context, ids []string) error {
	placeholders, args := buildPlaceholdersAndArgs(ids)
	query := fmt.Sprintf(`--sql
	update article_statuses
	set is_starred = case when article_id in (%s)
		then true
		else false
	end`, placeholders)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

type ArticleKind int

const (
	ArticleStarred ArticleKind = iota
	ArticleUnread
	ArticleAll
)

func (s *Sqlite) GetArticleByID(ctx context.Context, id string) (Article, error) {
	query := `--sql
	select a.id, a.title, a.href, a.content, a.author, s.is_read, s.is_starred, a.feed_id, f.title feed_name, a.published_at
	from articles a
	join article_statuses s on a.id = s.article_id
	join feeds f on f.id = a.feed_id
	where a.id = ?`

	var a Article
	if err := s.db.QueryRowContext(ctx, query, id).
		Scan(&a.ID, &a.Title, &a.Href, &a.Content, &a.Author, &a.IsRead, &a.IsStarred, &a.FeedID, &a.FeedTitle, &a.PublishedAt); err != nil {
		return a, err
	}

	return a, nil
}

type Article struct {
	ID          string
	Title       string
	Href        string
	Content     string
	Author      string
	IsRead      bool
	IsStarred   bool
	FeedID      string
	FeedTitle   string
	PublishedAt int64
}

var getArticlesWhereClause = map[ArticleKind]string{
	ArticleAll:     "",
	ArticleStarred: "where s.is_starred = true",
	ArticleUnread:  "where s.is_read = false",
}

func (s *Sqlite) GetArticles(ctx context.Context, kind ArticleKind) ([]Article, error) {
	query := fmt.Sprintf(`--sql
	select a.id, a.title, a.href, a.content, a.author, s.is_read, s.is_starred, a.feed_id, f.title feed_name, a.published_at
	from articles a
	join article_statuses s on a.id = s.article_id
	join feeds f on f.id = a.feed_id
	%s
	order by a.published_at desc`, getArticlesWhereClause[kind])

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Article
	for rows.Next() {
		var a Article
		if serr := rows.Scan(&a.ID, &a.Title, &a.Href, &a.Content, &a.Author, &a.IsRead, &a.IsStarred, &a.FeedID, &a.FeedTitle, &a.PublishedAt); serr != nil {
			return res, serr
		}

		res = append(res, a)
	}

	if err = rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func buildPlaceholdersAndArgs(in []string, prefixArgs ...any) (placeholders string, args []any) {
	placeholders = strings.Repeat("?,", len(in))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma

	args = make([]any, len(prefixArgs)+len(in))
	copy(args, prefixArgs)
	for i, v := range in {
		args[len(prefixArgs)+i] = v
	}

	return placeholders, args
}
