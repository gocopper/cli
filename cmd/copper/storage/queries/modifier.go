package queries

import (
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd, pkg string) error {
	return codemod.
		OpenDir(wd).
		Apply(codemod.CreateTemplateFiles(templatesFS, map[string]string{
			"pkg": pkg,
		}, false)).
		OpenFile(path.Join("./pkg", pkg, "wire.go")).
		Apply(codemod.AddProviderToWireSet("NewQueries")).
		CloseAndDone()
}
