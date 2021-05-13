package main

import (
	"context"
	"flag"
	"os"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdMigrate() *CmdMigrate {
	return &CmdMigrate{
		make: cli.NewMake(),
	}
}

type CmdMigrate struct {
	make *cli.Make
}

func (c *CmdMigrate) Name() string {
	return "migrate"
}

func (c *CmdMigrate) Synopsis() string {
	return "migrates database schemas"
}

func (c *CmdMigrate) Usage() string {
	return "copper migrate"
}

func (c *CmdMigrate) SetFlags(set *flag.FlagSet) {}

func (c *CmdMigrate) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	projectPath, err := os.Getwd()
	if err != nil {
		return 1
	}

	ok := c.make.Migrate(ctx, cli.MigrateParams{
		ProjectPath: projectPath,
	})
	if !ok {
		return 1
	}

	return 0
}
