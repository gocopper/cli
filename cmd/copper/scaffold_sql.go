package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/pkg/codemod/storage/sql"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewScaffoldSQLCmd(term *term.Terminal) *ScaffoldSQLCmd {
	return &ScaffoldSQLCmd{
		term: term,
	}
}

type ScaffoldSQLCmd struct {
	term *term.Terminal

	method string
	model  string
	field  string
	list   bool
}

func (c *ScaffoldSQLCmd) Name() string {
	return "scaffold:sql"
}

func (c *ScaffoldSQLCmd) Synopsis() string {
	return "Scaffolds SQL queries"
}

func (c *ScaffoldSQLCmd) Usage() string {
	return `copper scaffold:sql -method= -model= [-field=] [-list] foo
`
}

func (c *ScaffoldSQLCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.method, "method", "", "query, save")
	f.StringVar(&c.model, "model", "", "The database model for generating query")
	f.StringVar(&c.field, "field", "", "The field on the database model for generating query")
	f.BoolVar(&c.list, "list", false, "Set if the query should return a list of database models")
}

func (c *ScaffoldSQLCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	switch c.method {
	case "save":
		c.term.InProgressTask("Scaffold SQL")

		err := sql.NewSaveCodeMod(".", f.Arg(0), c.model).Apply(ctx)
		if err != nil {
			c.term.TaskFailed(err)
			return subcommands.ExitFailure
		}

		_ = mk.GoFmt(ctx, ".")

		c.term.TaskSucceeded()
	case "query":
		c.term.InProgressTask("Scaffold SQL")

		err := sql.NewQueryCodeMod(".", f.Arg(0), c.model, c.field, c.list).Apply(ctx)
		if err != nil {
			c.term.TaskFailed(err)
			return subcommands.ExitFailure
		}

		_ = mk.GoFmt(ctx, ".")

		c.term.TaskSucceeded()
	default:
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	return subcommands.ExitSuccess
}
