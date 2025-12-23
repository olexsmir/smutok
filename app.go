package main

import (
	"context"
	"errors"
	"log/slog"

	"olexsmir.xyz/smutok/internal/config"
	"olexsmir.xyz/smutok/internal/freshrss"
	"olexsmir.xyz/smutok/internal/store"
)

type app struct {
	cfg            *config.Config
	store          *store.Sqlite
	freshrss       *freshrss.Client
	freshrssSyncer *freshrss.Syncer
	freshrssWorker *freshrss.Worker
}

func bootstrap(ctx context.Context) (*app, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	store, err := store.NewSQLite(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	if merr := store.Migrate(ctx); merr != nil {
		return nil, merr
	}

	fr := freshrss.NewClient(cfg.FreshRSS.Host)
	token, err := getAuthToken(ctx, fr, store, cfg)
	if err != nil {
		return nil, err
	}
	fr.SetAuthToken(token)

	fs := freshrss.NewSyncer(fr, store)
	fw := freshrss.NewWorker()

	return &app{
		cfg:            cfg,
		store:          store,
		freshrss:       fr,
		freshrssSyncer: fs,
		freshrssWorker: fw,
	}, nil
}

func getAuthToken(ctx context.Context, fr *freshrss.Client, db *store.Sqlite, cfg *config.Config) (string, error) {
	token, err := db.GetToken(ctx)
	if err == nil {
		return token, nil
	}

	if !errors.Is(err, store.ErrNotFound) {
		return "", err
	}

	slog.Info("requesting auth key")

	token, err = fr.Login(ctx, cfg.FreshRSS.Username, cfg.FreshRSS.Password)
	if err != nil {
		return "", err
	}

	if serr := db.SetToken(ctx, token); serr != nil {
		return "", serr
	}

	return token, nil
}
