package store

import (
	"context"
	"fmt"
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
	placeholders, args := buildPlaceholdersAndArgs(currentFeedIDs)
	query := fmt.Sprintf(`--sql
	DELETE FROM feeds
	WHERE id NOT IN (%s)
	`, placeholders)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
