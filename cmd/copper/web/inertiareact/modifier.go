package inertiareact

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
		Cd("./web").
		Apply(
			codemod.RemoveFile("./src/App.tsx"),
			codemod.RemoveFile("./src/pages/index.html"),
			codemod.RunCmd("npm", "install", "@inertiajs/react@2"),
		).
		Cd("../").
		OpenFile("./cmd/app/wire.go").
		Apply(
			codemod.AddGoImports([]string{"github.com/gocopper/pkg/inertia"}),
			codemod.AddProviderToWireSet(`inertia.WireModule`),
		).
		CloseAndDone()
}
