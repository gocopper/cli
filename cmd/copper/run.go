package main

import (
	"context"
	"flag"
	"os"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdRun() *CmdRun {
	return &CmdRun{
		make: cli.NewMake(),
	}
}

type CmdRun struct {
	make *cli.Make
}

func (c *CmdRun) Name() string {
	return "run"
}

func (c *CmdRun) Synopsis() string {
	return "run the app"
}

func (c *CmdRun) Usage() string {
	return "copper run"
}

func (c *CmdRun) SetFlags(set *flag.FlagSet) {}

func (c *CmdRun) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	projectPath, err := os.Getwd()
	if err != nil {
		return 1
	}

	ok := c.make.Run(ctx, cli.RunParams{
		ProjectPath: projectPath,
		App:         true,
		JS:          true,
	})
	if !ok {
		return 1
	}

	return 0
}
