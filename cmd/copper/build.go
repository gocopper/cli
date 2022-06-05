package main

import (
	"context"
	"flag"

	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func NewBuildCmd(term *term.Terminal) *BuildCmd {
	return &BuildCmd{
		term: term,
	}
}

type BuildCmd struct {
	term *term.Terminal

	migrate bool
}

func (c *BuildCmd) Name() string {
	return "build"
}

func (c *BuildCmd) Synopsis() string {
	return "Builds the copper project"
}

func (c *BuildCmd) Usage() string {
	if mk.ProjectHasMigrate(".") {
		return `copper build [-migrate]
`
	}

	return `copper build
`
}

func (c *BuildCmd) SetFlags(f *flag.FlagSet) {
	if mk.ProjectHasMigrate(".") {
		f.BoolVar(&c.migrate, "migrate", true, "Build the migrate binary")
	}
}

func (c *BuildCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.term.InProgressTask("Build Project")

	err := mk.NewBuilder(".", c.migrate).Build(ctx)
	if err != nil {
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	c.term.TaskSucceeded()

	if c.migrate {
		c.term.Section("Run Database Migrations")
		c.term.Box("./build/migrate.out")
	}

	c.term.Section("Run App")
	c.term.Box("./build/app.out")

	return subcommands.ExitSuccess
}
