package store

import (
	"context"
	"fmt"
	"strings"
)

func (s *Sqlite) UpsertArticle(ctx context.Context, timestampUsec, feedID, title, content, author, href string, publishedAt int) error {
	tx, err := s.db.Begin()
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

// can be done like this?
// --sql
// update article_statuses
// set is_starred = case
// 	when article_id in (%s) then 1
// 	else 0
// end

func (s *Sqlite) SyncReadStatus(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		_, err := s.db.ExecContext(ctx, "update article_statuses set is_read = 1")
		return err
	}

	values := strings.Repeat("(?),", len(ids))
	values = values[:len(values)-1]

	args := make([]any, len(ids))
	for i, v := range ids {
		args[i] = v
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// make read those that are not in list
	readQuery := fmt.Sprintf(`--sql
	update article_statuses
	set is_read = true
	where article_id not in (%s)`, values)

	if _, err = tx.ExecContext(ctx, readQuery, args...); err != nil {
		return err
	}

	// make unread those that are in list
	unreadQuery := fmt.Sprintf(`--sql
	update article_statuses
	set is_read = false
	where article_id in (%s)`, values)

	if _, err = tx.ExecContext(ctx, unreadQuery, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Sqlite) SyncStarredStatus(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		_, err := s.db.ExecContext(ctx, "update article_statuses set is_starred = 0")
		return err
	}

	values := strings.Repeat("(?),", len(ids))
	values = values[:len(values)-1]

	args := make([]any, len(ids))
	for i, v := range ids {
		args[i] = v
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// make read those that are not in list
	readQuery := fmt.Sprintf(`--sql
	update article_statuses
	set is_starred = false
	where article_id not in (%s)`, values)

	if _, err = tx.ExecContext(ctx, readQuery, args...); err != nil {
		return err
	}

	// make unread those that are in list
	unreadQuery := fmt.Sprintf(`--sql
	update article_statuses
	set is_starred = true
	where article_id in (%s)`, values)

	if _, err = tx.ExecContext(ctx, unreadQuery, args...); err != nil {
		return err
	}

	return tx.Commit()
}
