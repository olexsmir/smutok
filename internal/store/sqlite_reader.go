package store

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Sqlite) GetLastSyncTime(ctx context.Context) (int64, error) {
	var lut int64
	err := s.db.QueryRowContext(ctx, "select last_sync from reader where id = 1 and last_sync is not null").Scan(&lut)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return lut, err
}

func (s *Sqlite) SetLastSyncTime(ctx context.Context, lastSync int64) error {
	_, err := s.db.ExecContext(ctx,
		`insert into reader (id, last_sync) values (1, ?)
		on conflict(id) do update set last_sync = excluded.last_sync`,
		lastSync)
	return err
}

func (s *Sqlite) GetToken(ctx context.Context) (string, error) {
	var tok string
	err := s.db.QueryRowContext(ctx, "select token from reader where id = 1 and token is not null").Scan(&tok)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return tok, err
}

func (s *Sqlite) SetToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx,
		`insert into reader (id, write_token) values (1, ?)
		on conflict(id) do update set token = excluded.token`,
		token)
	return err
}

func (s *Sqlite) GetWriteToken(ctx context.Context) (string, error) {
	var tok string
	err := s.db.QueryRowContext(ctx, "select write_token from reader where id = 1 and write_token is not null").Scan(&tok)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return tok, err
}

func (s *Sqlite) SetWriteToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx,
		`insert into reader (id, write_token) values (1, ?)
		on conflict(id) do update set write_token = excluded.write_token`,
		token)
	return err
}
