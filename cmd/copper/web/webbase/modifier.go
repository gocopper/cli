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
		indexPageRoute = codemod.InsertCHTTPRouteParams{
			Path:        "/",
			Method:      "Get",
			HandlerName: "HandleIndexPage",
			HandlerBody: `
	ro.html.WriteHTML(w, r, chttp.WriteHTMLParams{
		PageTemplate: "index.html",
	})`}
	)

	return codemod.
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		OpenFile("./cmd/app/wire.go").
		Apply(codemod.AddGoImports([]string{
			path.Join("{{.GoModule}}/web"),
			path.Join("{{.GoModule}}/web/build"),
		})).
		Apply(codemod.AddProviderToWireSet(webWireProviders)).
		CloseAndOpen("./pkg/app/router.go").
		Apply(
			codemod.AddGoImports([]string{"net/http"}),
			codemod.InsertLineAfter("type NewRouterParams struct {", "HTML *chttp.HTMLReaderWriter"),
			codemod.InsertLineAfter("return &Router{", "html: p.HTML,"),
			codemod.InsertLineAfter("type Router struct {", "html *chttp.HTMLReaderWriter"),
			codemod.InsertCHTTPRoute(indexPageRoute),
		).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.AppendText(chttpDevConfig)).
		CloseAndOpen("./pkg/app/handler.go").
		Apply(
			codemod.InsertLineAfter("*Router", "HTML *chttp.HTMLRouter"),
			codemod.InsertLineAfter("[]chttp.Router{", "p.HTML,"),
		).
		CloseAndDone()
}
