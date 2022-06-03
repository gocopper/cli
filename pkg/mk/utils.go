package mk

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
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
	cmd := exec.CommandContext(ctx, "wire", "gen", main)

	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
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
