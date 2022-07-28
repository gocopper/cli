package frontendnone

import (
	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var indexRoute = codemod.InsertCHTTPRouteParams{
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
		OpenDir(wd).
		OpenFile("./cmd/app/wire.go").
		Apply(codemod.AddProviderToWireSet("chttp.WireModuleEmptyHTML")).
		CloseAndOpen("./pkg/app/router.go").
		Apply(
			codemod.AddGoImports([]string{"net/http"}),
			codemod.InsertCHTTPRoute(indexRoute),
		).
		CloseAndDone()
}
