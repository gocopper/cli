package queries

import (
	"context"
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

//go:embed tmpl/*
var templatesFS embed.FS

func NewCodeMod(wd, pkg string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
		Pkg:        pkg,
	}
}

type CodeMod struct {
	WorkingDir string
	Pkg        string
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"pkg": cm.Pkg,
	}, false)
	if err != nil {
		return cerrors.New(err, "failed to create template files", nil)
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "pkg", cm.Pkg, "wire.go"), `
	NewQueries,
`)
	if err != nil {
		return cerrors.New(err, "failed to update wire.go", nil)
	}

	return nil
}
