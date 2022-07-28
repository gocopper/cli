package pkg

import (
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd, pkg string) error {
	var (
		pkgImportPath = path.Join("{{ .GoModule }}/pkg", pkg)
		pkgWireModule = pkg + ".WireModule"
	)

	return codemod.OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		Apply(codemod.CreateTemplateFiles(templatesFS, map[string]string{
			"pkg": pkg,
		}, false)).
		OpenFile("./pkg/app/wire.go").
		Apply(
			codemod.AddGoImports([]string{pkgImportPath}),
			codemod.AddProviderToWireSet(pkgWireModule),
		).
		CloseAndDone()
}
