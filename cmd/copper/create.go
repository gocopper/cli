package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gocopper/cli/cmd/copper/web/inertiareact"
	"path"
	"strings"

	"github.com/gocopper/cli/cmd/copper/app/appbase"
	"github.com/gocopper/cli/cmd/copper/storage/mysql"
	"github.com/gocopper/cli/cmd/copper/storage/postgres"
	"github.com/gocopper/cli/cmd/copper/storage/sqlite3"
	"github.com/gocopper/cli/cmd/copper/storage/storagebase"
	"github.com/gocopper/cli/cmd/copper/web/frontendnone"
	"github.com/gocopper/cli/cmd/copper/web/react"
	"github.com/gocopper/cli/cmd/copper/web/tailwind"
	"github.com/gocopper/cli/cmd/copper/web/tailwindvite"
	"github.com/gocopper/cli/cmd/copper/web/webbase"
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
	f.StringVar(&c.frontend, "frontend", "go", "go, go:tailwind, react, react:tailwind, inertia:react, inertia:react:tailwind, none")
	f.StringVar(&c.storage, "storage", "sqlite3", "sqlite3, postgres, mysql, none")
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

	err := appbase.Apply(wd, module)
	if err != nil {
		c.term.TaskFailed(cerrors.New(err, "failed to apply server code mod", nil))
		return subcommands.ExitFailure
	}

	if c.frontend != "none" {
		err = webbase.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply web code mod", nil))
			return subcommands.ExitFailure
		}
	}

	switch c.frontend {
	case "none":
		err = frontendnone.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply frontend none code mod", nil))
			return subcommands.ExitFailure
		}
	case "go":
		// note: already applied the web code mod
		break
	case "go:tailwind":
		err = tailwind.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply go:tailwind code mod", nil))
			return subcommands.ExitFailure
		}
	case "react":
		err = react.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply react code mod", nil))
			return subcommands.ExitFailure
		}
	case "react:tailwind":
		err = react.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply react code mod", nil))
			return subcommands.ExitFailure
		}

		err = tailwindvite.Apply(wd, false)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply tailwind (vite) code mod", nil))
			return subcommands.ExitFailure
		}
	case "inertia:react":
		err = react.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply react code mod", nil))
			return subcommands.ExitFailure
		}

		err = inertiareact.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply inertia (react) code mod", nil))
		}
	case "inertia:react:tailwind":
		err = react.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply react code mod", nil))
			return subcommands.ExitFailure
		}

		err = inertiareact.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply inertia (react) code mod", nil))
		}

		err = tailwindvite.Apply(wd, true)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply tailwind (vite) code mod", nil))
			return subcommands.ExitFailure
		}
	default:
		c.term.TaskFailed(errors.New("unknown frontend stack"))
		return subcommands.ExitUsageError
	}

	if c.storage != "none" {
		err = storagebase.Apply(wd)
		if err != nil {
			c.term.TaskFailed(cerrors.New(err, "failed to apply storage code mod", nil))
			return subcommands.ExitFailure
		}

		switch c.storage {
		case "sqlite3":
			err = sqlite3.Apply(wd)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply sqlite3 code mod", nil))
				return subcommands.ExitFailure
			}
		case "postgres":
			err = postgres.Apply(wd)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply postgres code mod", nil))
				return subcommands.ExitFailure
			}
		case "mysql":
			err = mysql.Apply(wd)
			if err != nil {
				c.term.TaskFailed(cerrors.New(err, "failed to apply mysql code mod", nil))
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
		packageManager := mk.GetPreferredPackageManager(path.Join(wd, "web"))
		c.term.Section("Run Tailwind Server")
		c.term.Box(fmt.Sprintf(`$ cd %s/web && %s run dev`, wd, packageManager))
	}

	if strings.Contains(c.frontend, "vite") {
		packageManager := mk.GetPreferredPackageManager(path.Join(wd, "web"))
		c.term.Section("Run Vite Server")
		c.term.Box(fmt.Sprintf(`$ cd %s/web && %s run dev`, wd, packageManager))
	}

	return subcommands.ExitSuccess
}
