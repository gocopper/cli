package main

import (
	"context"
	"errors"
	"flag"

	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewMigrateCmd(term *term.Terminal) *MigrateCmd {
	return &MigrateCmd{
		term: term,
	}
}

type MigrateCmd struct {
	term *term.Terminal
}

func (c *MigrateCmd) Name() string {
	return "migrate"
}

func (c *MigrateCmd) Synopsis() string {
	return "Runs database migrations"
}

func (c *MigrateCmd) Usage() string {
	return `copper migrate
`
}

func (c *MigrateCmd) SetFlags(f *flag.FlagSet) {}

func (c *MigrateCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.term.InProgressTask("Migrate Project")

	if !mk.ProjectHasMigrate(".") {
		c.term.TaskFailed(errors.New("project has no migrations"))
		return subcommands.ExitFailure
	}

	err := mk.NewBuilder(".", true).Build(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	c.term.TaskSucceeded()

	c.term.Section("Run Database Migrations")
	err = mk.NewRunner(".", "./build/migrate.out").Run(ctx)
	if err != nil {
		c.term.Error("Failed to run database migrations", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
