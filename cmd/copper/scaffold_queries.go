package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/cmd/copper/storage/queries"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldQueriesCmd(term *term.Terminal) *ScaffoldQueriesCmd {
	return &ScaffoldQueriesCmd{
		term: term,
	}
}

type ScaffoldQueriesCmd struct {
	term *term.Terminal
}

func (c *ScaffoldQueriesCmd) Name() string {
	return "scaffold:queries"
}

func (c *ScaffoldQueriesCmd) Synopsis() string {
	return "Scaffolds queries in a package"
}

func (c *ScaffoldQueriesCmd) Usage() string {
	return `copper scaffold:queries <package>
`
}

func (c *ScaffoldQueriesCmd) SetFlags(f *flag.FlagSet) {}

func (c *ScaffoldQueriesCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	c.term.InProgressTask("Scaffold queries")

	err := queries.NewCodeMod(".", f.Arg(0)).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	_ = mk.GoFmt(ctx, ".")

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
