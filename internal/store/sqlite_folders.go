package store

import "context"

func (s *Sqlite) UpsertTag(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `insert or replace into folders (id) values (?)`, id)
	return err
}
