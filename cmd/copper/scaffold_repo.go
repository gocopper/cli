package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/pkg/codemod/storage/repo"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldRepoCmd(term *term.Terminal) *ScaffoldRepoCmd {
	return &ScaffoldRepoCmd{
		term: term,
	}
}

type ScaffoldRepoCmd struct {
	term *term.Terminal
}

func (c *ScaffoldRepoCmd) Name() string {
	return "scaffold:repo"
}

func (c *ScaffoldRepoCmd) Synopsis() string {
	return "Scaffolds repository in a package"
}

func (c *ScaffoldRepoCmd) Usage() string {
	return `copper scaffold:repo foo
`
}

func (c *ScaffoldRepoCmd) SetFlags(f *flag.FlagSet) {}

func (c *ScaffoldRepoCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	c.term.InProgressTask("Scaffold storage repository")

	err := repo.NewCodeMod(".", f.Arg(0)).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	_ = mk.GoFmt(ctx, ".")

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
