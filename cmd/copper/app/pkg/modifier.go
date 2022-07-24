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

	return codemod.New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		CreateTemplateFiles(templatesFS, map[string]string{
			"pkg": pkg,
		}, false).
		OpenFile("./pkg/app/wire.go").
		Apply(
			codemod.ModAddGoImports([]string{pkgImportPath}),
			codemod.ModAddProviderToWireSet(pkgWireModule),
		).
		CloseAndDone()
}
