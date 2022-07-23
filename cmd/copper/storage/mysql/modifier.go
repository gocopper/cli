package mysql

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

func (cm *CodeMod) Apply(ctx context.Context) error {
	err := codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "base.toml"), `

[csql]
dialect="mysql"
`)
	if err != nil {
		return cerrors.New(err, "failed to add base csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "dev.toml"), `

[csql]
dsn="user:password@/dbname"
`)
	if err != nil {
		return cerrors.New(err, "failed to add dev csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/app/wire.go"), []string{
		"_ \"github.com/go-sql-driver/mysql\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add go-sql-driver/mysql import to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/migrate/wire.go"), []string{
		"_ \"github.com/go-sql-driver/mysql\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add go-sql-driver/mysql import to cmd/migrate/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
