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
	subcommands.Register(subcommands.HelpCommand(), "")

	subcommands.Register(NewScaffoldPkgCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldRepoCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldSQLCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldRouterCmd(terminal), "scaffold")
	subcommands.Register(NewScaffoldRouteCmd(terminal), "scaffold")

	flag.Parse()

	os.Exit(int(subcommands.Execute(ctx)))
}
