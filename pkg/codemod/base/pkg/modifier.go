package pkg

import (
	"context"
	"embed"
	"fmt"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

//go:embed tmpl/*
var templatesFS embed.FS

func NewCodeMod(wd, module, pkg string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
		Module:     module,
		Pkg:        pkg,
	}
}

type CodeMod struct {
	WorkingDir string
	Module     string
	Pkg        string
}

func (cm *CodeMod) Name() string {
	return "pkg"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"pkg": cm.Pkg,
	}, false)
	if err != nil {
		return cerrors.New(err, "failed to create template files", nil)
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg/app/wire.go"), []string{
		path.Join(cm.Module, "pkg", cm.Pkg),
	})
	if err != nil {
		return cerrors.New(err, "failed to add pkg imports to app/wire.go", nil)
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "pkg/app/wire.go"), fmt.Sprintf(`
%s.WireModule,
`, cm.Pkg))
	if err != nil {
		return cerrors.New(err, "failed to register pkg's wire module", nil)
	}

	return nil
}
