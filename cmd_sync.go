package main

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
)

var syncFeedsCmd = &cli.Command{
	Name:    "sync",
	Usage:   "Sync RSS feeds without opening the tui.",
	Aliases: []string{"s"},
	Action:  syncFeeds,
}

func syncFeeds(ctx context.Context, c *cli.Command) error {
	return errors.New("implement me")
}
