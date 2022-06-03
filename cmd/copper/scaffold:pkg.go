package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/v3/pkg/codemod"
	"github.com/gocopper/cli/v3/pkg/codemod/base/pkg"
	"github.com/gocopper/cli/v3/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldPkgCmd(term *term.Terminal) *ScaffoldPkgCmd {
	return &ScaffoldPkgCmd{
		term: term,
	}
}

type ScaffoldPkgCmd struct {
	term *term.Terminal
}

func (c *ScaffoldPkgCmd) Name() string {
	return "scaffold:pkg"
}

func (c *ScaffoldPkgCmd) Synopsis() string {
	return "Scaffolds a new package"
}

func (c *ScaffoldPkgCmd) Usage() string {
	return `copper scaffold:pkg foo
`
}

func (c *ScaffoldPkgCmd) SetFlags(f *flag.FlagSet) {}

func (c *ScaffoldPkgCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	c.term.InProgressTask("Scaffold package")

	module, err := codemod.GetGoModulePath(".")
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	err = pkg.NewCodeMod(".", module, f.Arg(0)).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
