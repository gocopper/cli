package main

import (
	"context"
	"flag"
	"os"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdBuild() *CmdBuild {
	return &CmdBuild{
		make: cli.NewMake(),
	}
}

type CmdBuild struct {
	make *cli.Make
}

func (c *CmdBuild) Name() string {
	return "build"
}

func (c *CmdBuild) Synopsis() string {
	return "build project binaries ready for deployment"
}

func (c *CmdBuild) Usage() string {
	return "copper build"
}

func (c *CmdBuild) SetFlags(set *flag.FlagSet) {}

func (c *CmdBuild) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	projectPath, err := os.Getwd()
	if err != nil {
		return 1
	}

	ok := c.make.Build(ctx, cli.BuildParams{
		ProjectPath: projectPath,
		Migrate:     true,
		JS:          true,
		App:         true,
	})
	if !ok {
		return 1
	}

	return 0
}
