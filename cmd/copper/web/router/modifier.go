package router

import (
	"context"
	"embed"
	"fmt"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
	"github.com/iancoleman/strcase"
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

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"pkg": cm.Pkg,
	}, false)
	if err != nil {
		return cerrors.New(err, "failed to create template files", nil)
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg/app/handler.go"), []string{
		path.Join(cm.Module, "pkg", cm.Pkg),
	})
	if err != nil {
		return cerrors.New(err, "failed to add imports to pkg/app/handler.go", nil)
	}

	err = codemod.ReplaceRegexInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"), "Logger +clogger\\.Logger", fmt.Sprintf(`%s

Logger clogger.Logger`, strcase.ToCamel(cm.Pkg)+" *"+cm.Pkg+".Router"))
	if err != nil {
		return cerrors.New(err, "failed to add router decl to pkg/app/handler.go", nil)
	}

	err = codemod.InsertLineAfterInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"), "Routers: []chttp.Router{", "p."+strcase.ToCamel(cm.Pkg)+",")
	if err != nil {
		return cerrors.New(err, "failed add router to []chttp.Router slice", nil)
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "pkg", cm.Pkg, "wire.go"), `
	wire.Struct(new(NewRouterParams), "*"),
	NewRouter,
`)
	if err != nil {
		return cerrors.New(err, "failed to update wire.go", nil)
	}

	return nil
}
