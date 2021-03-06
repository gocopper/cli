package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/cmd/copper/web/route"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
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
	return `copper scaffold:route -handler= -path= [-method=] <package>
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

	err := route.Apply(".", route.ApplyParams{
		Pkg:     f.Arg(0),
		Path:    c.path,
		Method:  c.method,
		Handler: c.handler,
	})
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	_ = mk.GoFmt(ctx, ".")

	c.term.TaskSucceeded()

	return subcommands.ExitSuccess
}
