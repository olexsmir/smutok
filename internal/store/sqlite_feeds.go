package store

import (
	"context"
	"fmt"
	"strings"
)

func (s *Sqlite) UpsertSubscription(ctx context.Context, id, title, url, htmlURL string) error {
	_, err := s.db.ExecContext(ctx,
		`insert or ignore into feeds (id, title, url, htmlUrl)
		values (?, ?, ?, ?)`,
		id, title, url, htmlURL)
	return err
}

func (s *Sqlite) LinkFeedWithFolder(ctx context.Context, feedID, folderID string) error {
	_, err := s.db.ExecContext(ctx,
		`insert or ignore into feed_folders (feed_id, folder_id)
		values (?, ?)`,
		feedID, folderID)
	return err
}

func (s *Sqlite) RemoveNonExistentFeeds(ctx context.Context, currentFeedIDs []string) error {
	if len(currentFeedIDs) == 0 {
		_, err := s.db.ExecContext(ctx, "delete from feeds")
		return err
	}

	values := strings.Repeat("(?),", len(currentFeedIDs))
	values = values[:len(values)-1] // trim trailing comma

	query := fmt.Sprintf(`--sql
	DELETE FROM feeds
	WHERE id NOT IN (VALUES %s)
	`, values)

	args := make([]any, len(currentFeedIDs))
	for i, v := range currentFeedIDs {
		args[i] = v
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
