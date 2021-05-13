package main

import (
	"context"
	"flag"
	"log"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdScaffold() *CmdScaffold {
	return &CmdScaffold{
		scaffold: cli.NewScaffold(),
	}
}

type CmdScaffold struct {
	scaffold *cli.Scaffold
}

func (c *CmdScaffold) Name() string {
	return "scaffold"
}

func (c *CmdScaffold) Synopsis() string {
	return "scaffold packages, routers, repositories.."
}

func (c *CmdScaffold) Usage() string {
	return "copper scaffold"
}

func (c *CmdScaffold) SetFlags(set *flag.FlagSet) {}

func (c *CmdScaffold) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		log.Fatalln(c.Usage())
	}

	var pkg = f.Arg(0)

	ok := c.scaffold.Run(ctx, pkg)
	if !ok {
		return 1
	}

	return 0
}
