package postgres

import (
	"context"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

func NewCodeMod(wd, module string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
		Module:     module,
	}
}

type CodeMod struct {
	WorkingDir string
	Module     string
}

func (cm *CodeMod) Name() string {
	return "postgres"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	err := codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "base.toml"), `

[csql]
dialect="postgres"
`)
	if err != nil {
		return cerrors.New(err, "failed to add base csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "dev.toml"), `

[csql]
dsn="user=postgres password=1234 host=127.0.0.1 port=5432 dbname=pg sslmode=disable"
`)
	if err != nil {
		return cerrors.New(err, "failed to add dev csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/app/wire.go"), []string{
		"_ \"github.com/lib/pq\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add lib/pq import to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/migrate/wire.go"), []string{
		"_ \"github.com/lib/pq\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add lib/pq import to cmd/migrate/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
