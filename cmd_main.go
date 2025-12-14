package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"

	"olexsmir.xyz/smutok/internal/config"
	"olexsmir.xyz/smutok/internal/provider"
	"olexsmir.xyz/smutok/internal/store"
	"olexsmir.xyz/smutok/internal/sync"
)

func runTui(ctx context.Context, c *cli.Command) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	db, err := store.NewSQLite(cfg.DBPath)
	if err != nil {
		return err
	}

	if merr := db.Migrate(ctx); merr != nil {
		return merr
	}

	gr := provider.NewFreshRSS(cfg.FreshRSS.Host)

	token, err := db.GetToken(ctx)
	if errors.Is(err, store.ErrNotFound) {
		slog.Info("authorizing")
		token, err = gr.Login(ctx, cfg.FreshRSS.Username, cfg.FreshRSS.Password)
		if err != nil {
			return err
		}

		if serr := db.SetToken(ctx, token); serr != nil {
			return serr
		}
	}
	if err != nil {
		return err
	}

	gr.SetAuthToken(token)

	gs := sync.NewFreshRSS(db, gr)
	fmt.Println(gs.Sync(ctx, true))

	return nil
}
