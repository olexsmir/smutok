package store

import (
	"context"
	"fmt"
	"strings"
)

type Action int

const (
	Read Action = iota
	Unread
	Star
	Unstar
)

func (a Action) String() string {
	switch a {
	case Read:
		return "read"
	case Unread:
		return "unread"
	case Star:
		return "star"
	case Unstar:
		return "unstar"
	default:
		return "unsupported"
	}
}

func (s *Sqlite) ChangeArticleStatus(ctx context.Context, articleID string, action Action) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update article status
	var query string
	switch action {
	case Read:
		query = `update article_statuses set is_read = 1 where article_id = ?`
	case Unread:
		query = `update article_statuses set is_read = 0 where article_id = ?`
	case Star:
		query = `update article_statuses set is_starred = 1 where article_id = ?`
	case Unstar:
		query = `update article_statuses set is_starred = 0 where article_id = ?`
	}

	e, err := tx.ExecContext(ctx, query, articleID)
	if err != nil {
		return err
	}
	if n, _ := e.RowsAffected(); n == 0 {
		return ErrNotFound
	}

	// enqueue action
	if _, err := tx.ExecContext(ctx, `insert into pending_actions (article_id, action) values (?, ?)`,
		articleID, action.String()); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Sqlite) GetPendingActions(ctx context.Context, action Action) ([]string, error) {
	query := `--sql
	select article_id
	from pending_actions
	where action = ?
	order by created_at desc
	limit 10`

	rows, err := s.db.QueryContext(ctx, query, action.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []string
	for rows.Next() {
		var pa string
		if serr := rows.Scan(&pa); serr != nil {
			return res, serr
		}
		res = append(res, pa)
	}

	if err = rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func (s *Sqlite) DeletePendingActions(ctx context.Context, action Action, articleIDs []string) error {
	placeholders := strings.Repeat("(?),", len(articleIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(articleIDs)+1)
	args[0] = action.String()
	for i, v := range articleIDs {
		args[i+1] = v
	}

	query := fmt.Sprintf(`--sql
	delete from pending_actions
	where action = ?
	  and article_id in (%s)
	`, placeholders)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
