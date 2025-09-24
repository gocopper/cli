package main

import (
	"context"
	"flag"
	"github.com/gocopper/cli/pkg/notifier"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"sync"
	"syscall"
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

	backgroundCmds []*exec.Cmd
	processLock    sync.Mutex
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
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

func (c *RunCmd) addBackgroundCmd(cmd *exec.Cmd) {
	c.processLock.Lock()
	defer c.processLock.Unlock()
	c.backgroundCmds = append(c.backgroundCmds, cmd)
}

func (c *RunCmd) killAllProcesses() {
	c.processLock.Lock()
	cmds := make([]*exec.Cmd, len(c.backgroundCmds))
	copy(cmds, c.backgroundCmds)
	c.backgroundCmds = nil
	c.processLock.Unlock()

	if len(cmds) == 0 {
		return
	}

	// Send SIGTERM to all processes
	for _, cmd := range cmds {
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}

	// Wait for all processes to exit with timeout
	done := make(chan bool, len(cmds))

	for _, cmd := range cmds {
		go func(c *exec.Cmd) {
			if c.Process != nil {
				c.Wait()
			}
			done <- true
		}(cmd)
	}

	// Wait for all processes or timeout after 10 seconds
	timeout := time.After(10 * time.Second)
	for i := 0; i < len(cmds); i++ {
		select {
		case <-done:
			// Process exited
		case <-timeout:
			// Force kill remaining processes
			for _, cmd := range cmds {
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}
			return
		}
	}
}

func (c *RunCmd) setupSignalHandling(ctx context.Context) context.Context {
	c.shutdownCtx, c.shutdownCancel = context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			c.term.Text("\nReceived signal: " + sig.String())
			c.killAllProcesses()
			c.shutdownCancel()
		case <-c.shutdownCtx.Done():
		}
	}()

	return c.shutdownCtx
}

func (c *RunCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	ctx = c.setupSignalHandling(ctx)
	defer c.killAllProcesses()

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
			c.killAllProcesses()
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
		case <-ctx.Done():
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
		packageManager := mk.GetPreferredPackageManager("./web")
		c.term.Section(packageManager + " run dev")
		runner := mk.NewBackgroundRunner(c.term, "./web", packageManager, "run", "dev")
		cmd, err := runner.RunBackground(ctx)
		if err != nil {
			c.term.Error("Failed to run '"+packageManager+" run dev'", err)
			return subcommands.ExitFailure
		}
		c.addBackgroundCmd(cmd)
	}

	if notify {
		notifier.Notify(notifier.NotifyParams{
			Title:       "Build Succeeded",
			Message:     "Server started",
			RemoveAfter: 3 * time.Second,
		})
	}

	c.term.Section("App Logs")
	runner := mk.NewBackgroundRunner(c.term, ".", "./build/app.out")
	cmd, err := runner.RunBackground(ctx)
	if err != nil {
		c.term.Error("Failed to run app", err)
		return subcommands.ExitFailure
	}
	c.addBackgroundCmd(cmd)

	err = cmd.Wait()
	if err != nil {
		c.term.Error("App exited with error", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
