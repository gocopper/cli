package mk

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/wire/pkg/wire"
)

func ProjectHasMigrate(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "cmd", "migrate"))

	return err == nil
}

func GoFmt(ctx context.Context, workingDir string) error {
	cmd := exec.CommandContext(ctx, "gofmt", "-w", ".")

	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func projectHasWeb(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "web", "src"))

	return err == nil
}

func goModTidy(ctx context.Context, workingDir string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")

	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func wireGen(ctx context.Context, workingDir, main string) error {
	outs, errs := wire.Generate(ctx,
		workingDir,
		os.Environ(),
		[]string{main},
		&wire.GenerateOptions{},
	)
	if len(errs) > 0 {
		return multiErrors(errs)
	} else if len(outs) == 0 {
		return errors.New("wire_gen.go was not generated")
	}

	for i := range outs {
		if len(outs[i].Errs) > 0 {
			return multiErrors(outs[i].Errs)
		}

		err := outs[i].Commit()
		if err != nil {
			return cerrors.New(err, "failed to write wire_gen.go", map[string]interface{}{
				"outputPath": outs[i].OutputPath,
			})
		}
	}

	return nil
}

func goBuild(ctx context.Context, workDir, main string) error {
	var (
		binary = path.Base(main) + ".out"
		out    = path.Join(workDir, "build", binary)

		cmd = exec.CommandContext(ctx, "go", "build", "-o", out, main)
	)

	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func multiErrors(errs []error) error {
	errMessages := make([]string, len(errs))
	for i := range errs {
		errMessages[i] = errs[i].Error()
	}

	return errors.New(strings.Join(errMessages, "\n\t"))
}
