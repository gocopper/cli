package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path"
	"strings"

	"github.com/gocopper/cli/pkg/codemod/base/server"
	"github.com/gocopper/cli/pkg/codemod/storage/gorm"
	"github.com/gocopper/cli/pkg/codemod/web/tailwind"
	"github.com/gocopper/cli/pkg/codemod/web/tailwindpostcss"
	"github.com/gocopper/cli/pkg/codemod/web/vitereact"
	"github.com/gocopper/cli/pkg/codemod/web/webgo"
	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/gocopper/copper/cerrors"
	"github.com/google/subcommands"
)

func NewCreateCmd(term *term.Terminal) *CreateCmd {
	return &CreateCmd{
		term: term,
	}
}

type CreateCmd struct {
	term *term.Terminal

	frontend string
	storage  string
}

func (c *CreateCmd) Name() string {
	return "create"
}

func (c *CreateCmd) Synopsis() string {
	return "Creates a new copper project"
}

func (c *CreateCmd) Usage() string {
	return `copper create [-frontend=] [-storage=] github.com/user/foo
`
}

func (c *CreateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.frontend, "frontend", "go", "go, go:tailwind, vite:react, vite:react:tailwind, none")
	f.StringVar(&c.storage, "storage", "gorm:sqlite", "gorm:sqlite, none")
}

func (c *CreateCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		c.term.Text(c.Usage())
		return subcommands.ExitUsageError
	}

	var (
		module = f.Arg(0)
		wd     = path.Join(".", path.Base(module))
	)

	c.term.InProgressTask(fmt.Sprintf("Create %s with frontend=%s, storage=%s", module, c.frontend, c.storage))

	err := server.NewCodeMod(wd, module).Apply(ctx)
	if err != nil {
		c.term.TaskFailed(cerrors.New(err, "failed to apply server code mod", nil))
		return subcommands.ExitFailure
	}

	if c.frontend != "none" {
		err = webgo.NewCodeMod(wd, module).Apply(ctx)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply web code mod", nil))
			return subcommands.ExitFailure
		}

		switch c.frontend {
		case "go":
			// note: already applied the web code mod
			break
		case "go:tailwind":
			err = tailwind.NewCodeMod(wd, module).Apply(ctx)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply go:tailwind code mod", nil))
				return subcommands.ExitFailure
			}
		case "vite:react":
			err = vitereact.NewCodeMod(wd, module).Apply(ctx)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply vite:react code mod", nil))
				return subcommands.ExitFailure
			}
		case "vite:react:tailwind":
			err = vitereact.NewCodeMod(wd, module).Apply(ctx)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply vite:react code mod", nil))
				return subcommands.ExitFailure
			}

			err = tailwindpostcss.NewCodeMod(wd, module).Apply(ctx)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply tailwind (postcss) code mod", nil))
				return subcommands.ExitFailure
			}
		default:
			c.term.TaskFailed(errors.New("unknown frontend stack"))
			return subcommands.ExitUsageError
		}
	}

	if c.storage != "none" {
		switch c.storage {
		case "gorm:sqlite":
			err = gorm.NewCodeMod(wd, module).Apply(ctx)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply gorm:sqlite code mod", nil))
				return subcommands.ExitFailure
			}
		default:
			c.term.TaskFailed(errors.New("unknown storage stack"))
			return subcommands.ExitUsageError
		}
	}

	_ = mk.GoFmt(ctx, ".")

	c.term.TaskSucceeded()

	c.term.Section("Run App Server")
	c.term.Box(fmt.Sprintf(`$ cd %s && copper run -watch`, wd))

	if c.frontend == "go:tailwind" {
		c.term.Section("Run Tailwind Server")
		c.term.Box(fmt.Sprintf(`$ cd %s/web && npm run dev`, wd))
	}

	if strings.Contains(c.frontend, "vite") {
		c.term.Section("Run Vite Server")
		c.term.Box(fmt.Sprintf(`$ cd %s/web && npm run dev`, wd))
	}

	return subcommands.ExitSuccess
}
