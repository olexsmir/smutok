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
