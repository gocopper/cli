package router

import (
	"embed"
	"fmt"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/iancoleman/strcase"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd, pkg string) error {
	var (
		routerDecl = fmt.Sprintf(`%s

Logger clogger.Logger`, strcase.ToCamel(pkg)+" *"+pkg+".Router")
		routerWireProviders = `wire.Struct(new(NewRouterParams), "*"),
	NewRouter`
	)

	return codemod.
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		Apply(codemod.CreateTemplateFiles(templatesFS, map[string]string{
			"pkg": pkg,
		}, false)).
		OpenFile("./pkg/app/handler.go").
		Apply(
			codemod.AddGoImports([]string{path.Join("{{.GoModule}}/pkg", pkg)}),
			codemod.ReplaceRegex("Logger +clogger\\.Logger", routerDecl),
			codemod.InsertLineAfter("Routers: []chttp.Router{", "p."+strcase.ToCamel(pkg)+","),
		).
		CloseAndOpen(path.Join("./pkg", pkg, "wire.go")).
		Apply(codemod.AddProviderToWireSet(routerWireProviders)).
		CloseAndDone()
}
