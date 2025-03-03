package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewVersionCmd(term *term.Terminal) *VersionCmd {
	return &VersionCmd{
		term: term,
	}
}

type VersionCmd struct {
	term *term.Terminal
}

func (c *VersionCmd) Name() string {
	return "version"
}

func (c *VersionCmd) Synopsis() string {
	return "Prints the current CLI version"
}

func (c *VersionCmd) Usage() string {
	return `copper version
`
}

func (c *VersionCmd) SetFlags(set *flag.FlagSet) {}

func (c *VersionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c.term.Text("v1.4.0")

	return subcommands.ExitSuccess
}
