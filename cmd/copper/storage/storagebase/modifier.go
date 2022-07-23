package storagebase

import (
	"context"
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
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

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"module": cm.Module,
	}, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), []string{
		"github.com/gocopper/copper/csql",
	})
	if err != nil {
		return cerrors.New(err, "failed to add csql import to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), `
csql.WireModule,
`)
	if err != nil {
		return cerrors.New(err, "failed to insert csql.WireModule to cmd/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg/app/handler.go"), []string{
		"github.com/gocopper/copper/csql",
	})
	if err != nil {
		return cerrors.New(err, "failed to add csql import to pkg/app/handler.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.InsertLineAfterInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"),
		"type NewHTTPHandlerParams struct {",
		`DatabaseTxMW *csql.TxMiddleware`,
	)
	if err != nil {
		return cerrors.New(err, "failed to add csql.TxMiddleware to pkg/app/handler.go", nil)
	}

	err = codemod.InsertLineAfterInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"),
		"[]chttp.Middleware{",
		`p.DatabaseTxMW,`,
	)
	if err != nil {
		return cerrors.New(err, "failed to add p.DatabaseTxMW to pkg/app/handler.go", nil)
	}

	return nil
}
