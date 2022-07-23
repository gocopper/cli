package mk

import (
	"context"
	_ "embed"
	"os"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
	"github.com/otiai10/copy"
)

//go:embed embed.go.tmpl
var embedGo string

func NewBuilder(wd string, migrate bool) *Builder {
	return &Builder{
		WorkingDir: wd,
		Migrate:    migrate,
	}
}

type Builder struct {
	WorkingDir string
	Migrate    bool
}

func (b *Builder) Build(ctx context.Context) error {
	if projectHasWeb(b.WorkingDir) {
		err := copy.Copy(path.Join(b.WorkingDir, "web", "public"), path.Join(b.WorkingDir, "web", "build", "static"))
		if err != nil {
			return cerrors.New(err, "failed to copy web static assets", nil)
		}

		err = os.WriteFile(path.Join(b.WorkingDir, "web/build/embed.go"), []byte(embedGo), 0644)
		if err != nil {
			return cerrors.New(err, "failed to write web/build/embed.go", nil)
		}
	}

	err := goModTidy(ctx, b.WorkingDir)
	if err != nil {
		return cerrors.New(err, "failed to run go mod tidy", nil)
	}

	module, err := codemod.GetGoModulePath(b.WorkingDir)
	if err != nil {
		return cerrors.New(err, "failed to get go module path", nil)
	}

	err = wireGen(ctx, b.WorkingDir, path.Join(module, "cmd", "app"))
	if err != nil {
		return cerrors.New(err, "failed to wire dependencies for cmd/app", nil)
	}

	err = goBuild(ctx, b.WorkingDir, path.Join(module, "cmd", "app"))
	if err != nil {
		return cerrors.New(err, "failed to build app binary", nil)
	}

	if b.Migrate {
		err := wireGen(ctx, b.WorkingDir, path.Join(module, "cmd", "migrate"))
		if err != nil {
			return cerrors.New(err, "failed to run wire dependencies for cmd/migrate", nil)
		}

		err = goBuild(ctx, b.WorkingDir, path.Join(module, "cmd", "migrate"))
		if err != nil {
			return cerrors.New(err, "failed to build migrate binary", nil)
		}
	}

	return nil
}
