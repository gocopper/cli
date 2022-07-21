package sqlite3

import (
	"context"
	"fmt"
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
	return "sqlite3"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	err := codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "base.toml"), fmt.Sprintf(`

[csql]
dialect="sqlite3"
dsn="./%s.db"
`, path.Base(cm.Module)))
	if err != nil {
		return cerrors.New(err, "failed to add csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/app/wire.go"), []string{
		"_ \"github.com/mattn/go-sqlite3\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add go-sqlite3 import to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd/migrate/wire.go"), []string{
		"_ \"github.com/mattn/go-sqlite3\"",
	})
	if err != nil {
		return cerrors.New(err, "failed to add go-sqlite3 import to cmd/migrate/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
