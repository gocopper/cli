package gorm

import (
	"context"
	"embed"
	"fmt"
	"path"

	codemod2 "github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

//go:embed tmpl/*
var templatesFS embed.FS

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
	return "GORM"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod2.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"module": cm.Module,
	}, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod2.AppendTextToFile(path.Join(cm.WorkingDir, "config", "base.toml"), fmt.Sprintf(`

[csql]
dialect="sqlite"
dsn="./%s.db"
`, path.Base(cm.Module)))
	if err != nil {
		return cerrors.New(err, "failed to add csql config", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod2.AddImports(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), []string{
		"github.com/gocopper/copper/csql",
	})
	if err != nil {
		return cerrors.New(err, "failed to add csql import to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod2.InsertWireModuleItems(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), `
csql.WireModule,
`)
	if err != nil {
		return cerrors.New(err, "failed to insert csql.WireModule to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
