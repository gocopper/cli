package storagebase

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	return codemod.
		New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		CreateTemplateFiles(templatesFS, nil, true).
		OpenFile("./cmd/app/wire.go").
		Apply(
			codemod.ModAddGoImports([]string{"github.com/gocopper/copper/csql"}),
			codemod.ModAddProviderToWireSet("csql.WireModule"),
		).
		CloseAndOpen("./pkg/app/handler.go").
		Apply(
			codemod.ModAddGoImports([]string{"github.com/gocopper/copper/csql"}),
			codemod.ModInsertLineAfter("type NewHTTPHandlerParams struct {\n", "DatabaseTxMW *csql.TxMiddleware"),
			codemod.ModInsertLineAfter("[]chttp.Middleware{\n", "p.DatabaseTxMW,"),
		).
		CloseAndDone()
}
