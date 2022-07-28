package storagebase

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	return codemod.
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		OpenFile("./cmd/app/wire.go").
		Apply(
			codemod.AddGoImports([]string{"github.com/gocopper/copper/csql"}),
			codemod.AddProviderToWireSet("csql.WireModule"),
		).
		CloseAndOpen("./pkg/app/handler.go").
		Apply(
			codemod.AddGoImports([]string{"github.com/gocopper/copper/csql"}),
			codemod.InsertLineAfter("type NewHTTPHandlerParams struct {\n", "DatabaseTxMW *csql.TxMiddleware"),
			codemod.InsertLineAfter("[]chttp.Middleware{\n", "p.DatabaseTxMW,"),
		).
		CloseAndDone()
}
