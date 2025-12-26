package store

import (
	"context"
	"database/sql"
	_ "embed"
	"log/slog"

	amigrate "ariga.io/atlas/sql/migrate"
	aschema "ariga.io/atlas/sql/schema"
	asqlite "ariga.io/atlas/sql/sqlite"

	_ "modernc.org/sqlite"
)

//go:embed schema.hcl
var schema []byte

type Sqlite struct {
	db *sql.DB
}

func NewSQLite(path string) (*Sqlite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &Sqlite{
		db: db,
	}, nil
}

func (s *Sqlite) Close() error { return s.db.Close() }

func (s *Sqlite) Migrate(ctx context.Context) error {
	driver, err := asqlite.Open(s.db)
	if err != nil {
		return err
	}

	want := &aschema.Schema{}
	if serr := asqlite.EvalHCLBytes(schema, want, nil); serr != nil {
		return err
	}

	got, err := driver.InspectSchema(ctx, "", nil)
	if err != nil {
		return err
	}

	changes, err := driver.SchemaDiff(got, want)
	if err != nil {
		return err
	}

	slog.Debug("running migration")
	if merr := driver.ApplyChanges(ctx, changes, []amigrate.PlanOption{}...); merr != nil {
		return merr
	}

	_, err = driver.ExecContext(ctx, `--sql
		PRAGMA foreign_keys = ON`)
	return err
}
