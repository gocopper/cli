package main

import (
	"context"
	"flag"
	"os"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdWatch() *CmdWatch {
	return &CmdWatch{
		make: cli.NewMake(),
	}
}

type CmdWatch struct {
	make *cli.Make
}

func (c *CmdWatch) Name() string {
	return "watch"
}

func (c *CmdWatch) Synopsis() string {
	return "run and watch the app"
}

func (c *CmdWatch) Usage() string {
	return "copper watch"
}

func (c *CmdWatch) SetFlags(set *flag.FlagSet) {}

func (c *CmdWatch) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	projectPath, err := os.Getwd()
	if err != nil {
		return 1
	}

	ok := c.make.Watch(ctx, cli.WatchParams{
		ProjectPath: projectPath,
	})
	if !ok {
		return 1
	}

	return 0
}
