package frontendnone

import (
	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var indexRoute = codemod.ModInsertCHTTPRouteParams{
		Path:        "/",
		Method:      "Get",
		HandlerName: "HandleIndex",
		HandlerBody: `
	ro.rw.WriteJSON(w, chttp.WriteJSONParams{
		Data: map[string]string{
			"message": "Hello, Copper!",
			"demo": "https://vimeo.com/723537998",
			"frontend_stack": "none",
		},
	})`,
	}

	return codemod.
		New(wd).
		OpenFile("./cmd/app/wire.go").
		Apply(codemod.ModAddProviderToWireSet("chttp.WireModuleEmptyHTML")).
		CloseAndOpen("./pkg/app/router.go").
		Apply(
			codemod.ModAddGoImports([]string{"net/http"}),
			codemod.ModInsertCHTTPRoute(indexRoute),
		).
		CloseAndDone()
}
