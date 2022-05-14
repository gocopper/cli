package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(NewCmdInit(), "")
	subcommands.Register(NewCmdBuild(), "")
	subcommands.Register(NewCmdMigrate(), "")
	subcommands.Register(NewCmdRun(), "")
	subcommands.Register(NewCmdWatch(), "")
	subcommands.Register(NewCmdScaffold(), "")
	subcommands.Register(NewCmdVersion(), "")

	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	var (
		ctx      = context.Background()
		exitCode = subcommands.Execute(ctx)
	)

	os.Exit(int(exitCode))
}
