package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"olexsmir.xyz/smutok/internal/config"
)

//go:embed version
var _version string

var version = strings.Trim(_version, "\n")

func main() {
	cmd := &cli.Command{
		Name:                  "smutok",
		Version:               version,
		Usage:                 "An RSS feed reader.",
		EnableShellCompletion: true,
		Action:                runTui,
		Commands: []*cli.Command{
			initConfigCmd,
			syncFeedsCmd,
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runTui(ctx context.Context, c *cli.Command) error {
	return errors.New("there's no tui, i lied")
}

// sync

var syncFeedsCmd = &cli.Command{
	Name:    "sync",
	Usage:   "Sync RSS feeds without opening the tui.",
	Aliases: []string{"s"},
	Action:  syncFeeds,
}

func syncFeeds(ctx context.Context, c *cli.Command) error {
	app, err := bootstrap(ctx)
	if err != nil {
		return err
	}
	return app.freshrssSyncer.Sync(ctx)
}

// init

var initConfigCmd = &cli.Command{
	Name:   "init",
	Usage:  "Initialize smutok's config",
	Action: initConfig,
}

func initConfig(ctx context.Context, c *cli.Command) error {
	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}
	slog.Info("Config was initialized, enter your credentials", "file", config.MustGetConfigFilePath())
	return nil
}
