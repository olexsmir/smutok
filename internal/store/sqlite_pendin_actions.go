package store

import (
	"context"
	"fmt"
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

var changeArticleStatusQuery = map[Action]string{
	Read:   `update article_statuses set is_read = 1 where article_id = ?`,
	Unread: `update article_statuses set is_read = 0 where article_id = ?`,
	Star:   `update article_statuses set is_starred = 1 where article_id = ?`,
	Unstar: `update article_statuses set is_starred = 0 where article_id = ?`,
}

func (s *Sqlite) ChangeArticleStatus(ctx context.Context, articleID string, action Action) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update article status
	e, err := tx.ExecContext(ctx, changeArticleStatusQuery[action], articleID)
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

func (s *Sqlite) DeletePendingActions(
	ctx context.Context,
	action Action,
	articleIDs []string,
) error {
	placeholders, args := buildPlaceholdersAndArgs(articleIDs, action.String())
	query := fmt.Sprintf(`--sql
	delete from pending_actions
	where action = ?
	  and article_id in (%s)
	`, placeholders)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
