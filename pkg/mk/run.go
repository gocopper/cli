package mk

import (
	"context"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/gocopper/copper/cerrors"
)

func NewRunner(wd, binary string) *Runner {
	return &Runner{
		WorkingDir: wd,
		Binary:     binary,
	}
}

type Runner struct {
	WorkingDir string
	Binary     string
}

func (r *Runner) Run(ctx context.Context) error {
	const SignalKilled = 9

	cmd := exec.CommandContext(ctx, path.Join(r.WorkingDir, "build", r.Binary))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return cerrors.New(err, "cmd did not exit cleanly", nil)
		}

		status := exitErr.Sys().(syscall.WaitStatus)
		if status.Signaled() && status.Signal() != SignalKilled {
			return cerrors.New(err, "cmd did not exit cleanly", nil)
		}
	}

	return nil
}
