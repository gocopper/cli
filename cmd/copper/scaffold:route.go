package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/v3/pkg/codemod/web/route"
	"github.com/gocopper/cli/v3/pkg/mk"
	"github.com/gocopper/cli/v3/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldRouteCmd(term *term.Terminal) *ScaffoldRouteCmd {
	return &ScaffoldRouteCmd{
		term: term,
	}
}

type ScaffoldRouteCmd struct {
	term *term.Terminal

	handler string
	path    string
	method  string
}

func (c *ScaffoldRouteCmd) Name() string {
	return "scaffold:route"
}

func (c *ScaffoldRouteCmd) Synopsis() string {
	return "Scaffolds a route in a router"
}

func (c *ScaffoldRouteCmd) Usage() string {
	return `copper scaffold:route -handler= -path= [-method=] foo
`
}

func (c *ScaffoldRouteCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.handler, "handler", "", "The new route handler's name (ex. HandleIndexPage)")
	f.StringVar(&c.path, "path", "", "The URL path that invokes the handler (ex. /books/{id})")
	f.StringVar(&c.method, "method", "Get", "Get, Post, Put, Patch, Delete")
}

func (c *ScaffoldRouteCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	c.term.InProgressTask("Scaffold route")

	err := route.NewCodeMod(".", f.Arg(0), c.path, c.method, c.handler).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	_ = mk.GoLangCILint(ctx, ".")

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
