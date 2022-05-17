package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
)

const version = "0.6.1"

func NewCmdVersion() *CmdVersion {
	return &CmdVersion{}
}

type CmdVersion struct{}

func (cmd *CmdVersion) Name() string {
	return "version"
}

func (cmd *CmdVersion) Synopsis() string {
	return "version synopsis"
}

func (cmd *CmdVersion) Usage() string {
	return "version usage"
}

func (cmd *CmdVersion) SetFlags(f *flag.FlagSet) {}

func (cmd *CmdVersion) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	fmt.Println("Version ", version)

	return 0
}
