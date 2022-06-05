package webgo

import (
	"context"
	"embed"
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
	return "go"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod2.CreateTemplateFiles(templatesFS, cm.WorkingDir, nil, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", nil)
	}

	err = codemod2.AddImports(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), []string{
		path.Join(cm.Module, "web"),
		path.Join(cm.Module, "web", "build"),
	})
	if err != nil {
		return cerrors.New(err, "failed to add web imports to cmd/app/wire.go", nil)
	}

	err = codemod2.InsertWireModuleItems(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), `
wire.InterfaceValue(new(chttp.HTMLDir), web.HTMLDir),
wire.InterfaceValue(new(chttp.StaticDir), build.StaticDir),
web.HTMLRenderFuncs,`)
	if err != nil {
		return cerrors.New(err, "failed to add web wire modules to cmd/app/wire.go", nil)
	}

	err = codemod2.AddImports(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), []string{
		"net/http",
	})
	if err != nil {
		return cerrors.New(err, "failed to add net/http import to pkg/app/router.go", nil)
	}

	err = codemod2.InsertRoute(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), codemod2.InsertRouteParams{
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

	err = codemod2.AppendTextToFile(path.Join(cm.WorkingDir, "config", "dev.toml"), `
[chttp]
use_local_html = true
render_html_error = true
`)
	if err != nil {
		return cerrors.New(err, "failed to add chttp config to dev.toml", nil)
	}

	err = insertHTMLRouterToAppHandler(path.Join(cm.WorkingDir, "pkg", "app", "handler.go"))
	if err != nil {
		return cerrors.New(err, "failed to register chttp.HTMLRouter", nil)
	}

	return nil
}
