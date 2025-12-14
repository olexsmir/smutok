package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
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
