package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli"
	"github.com/google/subcommands"
)

func NewCmdInit() *CmdInit {
	return &CmdInit{
		scaffold: cli.NewScaffold(),
	}
}

type CmdInit struct {
	scaffold *cli.Scaffold
}

func (cmd *CmdInit) Name() string {
	return "init"
}

func (cmd *CmdInit) Synopsis() string {
	return "init synopsis"
}

func (cmd *CmdInit) Usage() string {
	return "init usage"
}

func (cmd *CmdInit) SetFlags(f *flag.FlagSet) {}

func (cmd *CmdInit) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	ok := cmd.scaffold.Init()
	if !ok {
		return 1
	}

	return 0
}
