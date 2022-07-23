package webgo

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
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, nil, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", nil)
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), []string{
		path.Join(cm.Module, "web"),
		path.Join(cm.Module, "web", "build"),
	})
	if err != nil {
		return cerrors.New(err, "failed to add web imports to cmd/app/wire.go", nil)
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), `
wire.InterfaceValue(new(chttp.HTMLDir), web.HTMLDir),
wire.InterfaceValue(new(chttp.StaticDir), build.StaticDir),
web.HTMLRenderFuncs,`)
	if err != nil {
		return cerrors.New(err, "failed to add web wire modules to cmd/app/wire.go", nil)
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), []string{
		"net/http",
	})
	if err != nil {
		return cerrors.New(err, "failed to add net/http import to pkg/app/router.go", nil)
	}

	err = codemod.InsertRoute(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), codemod.InsertRouteParams{
		Path:        "/",
		Method:      "Get",
		HandlerName: "HandleIndexPage",
		HandlerBody: `
	ro.rw.WriteHTML(w, r, chttp.WriteHTMLParams{
		PageTemplate: "index.html",
	})`,
	})
	if err != nil {
		return cerrors.New(err, "failed to insert route for index page", nil)
	}

	err = codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "dev.toml"), `
[chttp]
use_local_html = true
render_html_error = true
`)
	if err != nil {
		return cerrors.New(err, "failed to add chttp config to dev.toml", nil)
	}

	err = codemod.InsertLineAfterInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"), "*Router", "HTML *chttp.HTMLRouter")
	if err != nil {
		return cerrors.New(err, "failed to insert *chttp.HTMLRouter", nil)
	}

	err = codemod.InsertLineAfterInFile(path.Join(cm.WorkingDir, "pkg/app/handler.go"), "p.App,", "p.HTML,")
	if err != nil {
		return fmt.Errorf("failed to insert p.HTML; %v", err)
	}

	return nil
}
