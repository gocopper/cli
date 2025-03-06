package mk

import (
	"context"
	"errors"
	"github.com/gocopper/cli/pkg/term"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gocopper/copper/cerrors"
)

func NewRunner(wd, binary string, args ...string) *Runner {
	return &Runner{
		WorkingDir: wd,
		Binary:     binary,
		Args:       args,
	}
}

func NewBackgroundRunner(term *term.Terminal, wd, binary string, args ...string) *Runner {
	return &Runner{
		WorkingDir: wd,
		Binary:     binary,
		Args:       args,
		Background: true,
		Term:       term,
	}
}

type Runner struct {
	WorkingDir string
	Binary     string
	Args       []string
	Background bool
	Term       *term.Terminal
}

func (r *Runner) Run(ctx context.Context) error {
	const SignalKilled = 9

	cmd := exec.CommandContext(ctx, r.Binary, r.Args...)
	cmd.Dir = r.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if r.Background {
		if err := cmd.Start(); err != nil {
			return cerrors.New(err, "failed to start background command", nil)
		}

		outputCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		go func() {
			err := cmd.Wait()
			if err != nil {
				var exitErr *exec.ExitError
				if ok := errors.As(err, &exitErr); !ok {
					r.Term.Error("Background cmd did not exit cleanly", err)
					return
				}

				status := exitErr.Sys().(syscall.WaitStatus)
				if status.Signaled() && status.Signal() != SignalKilled {
					r.Term.Error("Background cmd exited with signal: "+status.Signal().String(), nil)
				}
			}
		}()

		<-outputCtx.Done()

		if errors.Is(outputCtx.Err(), context.DeadlineExceeded) {
			cmd.Stdout = nil
			cmd.Stderr = nil
		}

		return nil
	}

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); !ok {
			return cerrors.New(err, "cmd did not exit cleanly", nil)
		}

		status := exitErr.Sys().(syscall.WaitStatus)
		if status.Signaled() && status.Signal() != SignalKilled {
			return cerrors.New(err, "cmd did not exit cleanly", nil)
		}
	}

	return nil
}
