package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/pkg/codemod/web/router"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldRouterCmd(term *term.Terminal) *ScaffoldRouterCmd {
	return &ScaffoldRouterCmd{
		term: term,
	}
}

type ScaffoldRouterCmd struct {
	term *term.Terminal
}

func (c *ScaffoldRouterCmd) Name() string {
	return "scaffold:router"
}

func (c *ScaffoldRouterCmd) Synopsis() string {
	return "Scaffolds a router"
}

func (c *ScaffoldRouterCmd) Usage() string {
	return `copper scaffold:router foo
`
}

func (c *ScaffoldRouterCmd) SetFlags(f *flag.FlagSet) {}

func (c *ScaffoldRouterCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	c.term.InProgressTask("Scaffold router")

	err := router.NewCodeMod(".", f.Arg(0)).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	_ = mk.GoFmt(ctx, ".")

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
