package route

import (
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

type ApplyParams struct {
	Pkg     string
	Path    string
	Method  string
	Handler string
}

func Apply(wd string, p ApplyParams) error {
	return codemod.
		New(wd).
		OpenFile(path.Join("./pkg", p.Pkg, "router.go")).
		Apply(
			codemod.ModAddGoImports([]string{"net/http"}),
			codemod.ModInsertCHTTPRoute(codemod.ModInsertCHTTPRouteParams{
				Path:        p.Path,
				Method:      p.Method,
				HandlerName: p.Handler,
				HandlerBody: "",
			}),
		).
		CloseAndDone()
}
