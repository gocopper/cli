package frontendnone

import (
	"embed"
	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	return codemod.
		OpenDir(wd).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		OpenFile("./cmd/app/wire.go").
		Apply(codemod.AddProviderToWireSet("chttp.WireModuleEmptyHTML")).
		CloseAndOpen("./pkg/app/router.go").
		CloseAndDone()
}
