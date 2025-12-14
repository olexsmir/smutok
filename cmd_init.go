package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"olexsmir.xyz/smutok/internal/config"
)

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
