package webbase

import (
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	var (
		webWireProviders = `
wire.InterfaceValue(new(chttp.HTMLDir), web.HTMLDir),
wire.InterfaceValue(new(chttp.StaticDir), build.StaticDir),
web.HTMLRenderFuncs`
		chttpDevConfig = `
[chttp]
use_local_html = true
render_html_error = true
`
		indexPageRoute = codemod.ModInsertCHTTPRouteParams{
			Path:        "/",
			Method:      "Get",
			HandlerName: "HandleIndexPage",
			HandlerBody: `
	ro.rw.WriteHTML(w, r, chttp.WriteHTMLParams{
		PageTemplate: "index.html",
	})`}
	)

	return codemod.
		New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		CreateTemplateFiles(templatesFS, nil, true).
		OpenFile("./cmd/app/wire.go").
		Apply(codemod.ModAddGoImports([]string{
			path.Join("{{.GoModule}}/web"),
			path.Join("{{.GoModule}}/web/build"),
		})).
		Apply(codemod.ModAddProviderToWireSet(webWireProviders)).
		CloseAndOpen("./pkg/app/router.go").
		Apply(
			codemod.ModAddGoImports([]string{"net/http"}),
			codemod.ModInsertCHTTPRoute(indexPageRoute),
		).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.ModAppendText(chttpDevConfig)).
		CloseAndOpen("./pkg/app/handler.go").
		Apply(
			codemod.ModInsertLineAfter("*Router", "HTML *chttp.HTMLRouter"),
			codemod.ModInsertLineAfter("[]chttp.Router{", "p.HTML,"),
		).
		CloseAndDone()
}
