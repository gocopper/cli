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
		OpenDir(wd).
		OpenFile(path.Join("./pkg", p.Pkg, "router.go")).
		Apply(
			codemod.AddGoImports([]string{"net/http"}),
			codemod.InsertCHTTPRoute(codemod.InsertCHTTPRouteParams{
				Path:        p.Path,
				Method:      p.Method,
				HandlerName: p.Handler,
				HandlerBody: "",
			}),
		).
		CloseAndDone()
}
