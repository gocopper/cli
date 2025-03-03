package main

import (
	"context"
	"flag"
	"os"

	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
)

func main() {
	var (
		ctx      = context.Background()
		terminal = term.NewTerminal()
	)

	subcommands.Register(NewCreateCmd(terminal), "")
	subcommands.Register(NewBuildCmd(terminal), "")
	subcommands.Register(NewRunCmd(terminal), "")
	subcommands.Register(NewMigrateCmd(terminal), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(NewVersionCmd(terminal), "")

	subcommands.Register(NewScaffoldPkgCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldQueriesCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldRouterCmd(terminal), "scaffold")

	flag.Parse()

	os.Exit(int(subcommands.Execute(ctx)))
}
