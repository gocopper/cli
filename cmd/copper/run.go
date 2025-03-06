package main

import (
	"context"
	"flag"
	"github.com/gocopper/cli/pkg/notifier"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/gocopper/cli/pkg/mk"
	"github.com/gocopper/cli/pkg/term"
	"github.com/google/subcommands"
	"github.com/radovskyb/watcher"
)

func NewRunCmd(term *term.Terminal) *RunCmd {
	return &RunCmd{
		term:       term,
		isFirstRun: true,
	}
}

type RunCmd struct {
	term *term.Terminal

	migrate bool
	npm     bool
	watch   bool

	isFirstRun     bool
	isFirstRunOnce sync.Once
}

func (c *RunCmd) Name() string {
	return "run"
}

func (c *RunCmd) Synopsis() string {
	return "Runs the copper project"
}

func (c *RunCmd) Usage() string {
	if mk.ProjectHasMigrate(".") {
		return `copper run [-migrate] [-watch]
`
	}

	return `copper run [-watch]
`
}

func (c *RunCmd) SetFlags(f *flag.FlagSet) {
	if mk.ProjectHasMigrate(".") {
		f.BoolVar(&c.migrate, "migrate", true, "Run database migrations")
	}

	if mk.ProjectHasWeb(".") {
		f.BoolVar(&c.npm, "npm", true, "Run 'npm run dev'")
	}

	f.BoolVar(&c.watch, "watch", false, "Automatically restart project on source changes")
}

func (c *RunCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if !c.watch {
		return c.execute(ctx)
	}

	w := watcher.New()

	w.SetMaxEvents(1)

	err := w.AddRecursive(path.Join(".", "pkg"))
	if err != nil {
		c.term.Error("Failed to watch pkg dir", err)
		return subcommands.ExitFailure
	}

	w.AddFilterHook(watcher.RegexFilterHook(regexp.MustCompile(".*.go$"), false))

	if _, err := os.Stat(path.Join(".", "web/wire.go")); err == nil {
		err = w.Add(path.Join(".", "web/wire.go"))
		if err != nil {
			c.term.Error("Failed to watch web/wire.go", err)
			return subcommands.ExitFailure
		}
	}

	go func() {
		err := w.Start(time.Millisecond * 500)
		if err != nil {
			c.term.Error("Failed to start watching pkg dir", err)
		}
	}()

	runCtx, cancelRun := context.WithCancel(ctx)

	for {
		select {
		case <-w.Event:
			if !c.isFirstRun {
				c.term.Text("\n------------------------------------------------------------------------")

				notifier.Notify(notifier.NotifyParams{
					Title:   "File Changed",
					Message: "Restarting server..",
				})
			}

			cancelRun()
			runCtx.Done()

			runCtx, cancelRun = context.WithCancel(ctx)
			go func() {
				c.execute(runCtx)
			}()
		case err := <-w.Error:
			cancelRun()
			c.term.Error("Error while watching pkg dir", err)
			return subcommands.ExitFailure
		case <-w.Closed:
			cancelRun()
			return subcommands.ExitSuccess
		}
	}
}

func (c *RunCmd) execute(ctx context.Context) subcommands.ExitStatus {
	c.term.InProgressTask("Build Project")

	migrate := c.migrate && c.isFirstRun
	npm := c.npm && c.isFirstRun
	notify := !c.isFirstRun

	c.isFirstRunOnce.Do(func() {
		c.isFirstRun = false
	})

	err := mk.NewBuilder(".", migrate).Build(ctx)
	if err != nil {
		notifier.Notify(notifier.NotifyParams{
			Title:   "Build Failed",
			Message: err.Error(),
		})
		c.term.TaskFailed(err)
		return subcommands.ExitFailure
	}

	c.term.TaskSucceeded()

	if migrate {
		c.term.Section("Run Database Migrations")
		err := mk.NewRunner(".", "./build/migrate.out", "-set", "csql.migrations.source=\"dir\"").Run(ctx)
		if err != nil {
			c.term.Error("Failed to run database migrations", err)
			return subcommands.ExitFailure
		}
	}

	if npm {
		c.term.Section("npm run dev")
		err := mk.NewBackgroundRunner(c.term, "./web", "npm", "run", "dev").Run(ctx)
		if err != nil {
			c.term.Error("Failed to run 'npm run dev'", err)
			return subcommands.ExitFailure
		}
	}

	if notify {
		notifier.Notify(notifier.NotifyParams{
			Title:       "Build Succeeded",
			Message:     "Server started",
			RemoveAfter: 3 * time.Second,
		})
	}

	c.term.Section("App Logs")
	err = mk.NewRunner(".", "./build/app.out").Run(ctx)
	if err != nil {
		c.term.Error("Failed to run app", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
