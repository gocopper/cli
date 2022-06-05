package route

import (
	"context"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

func NewCodeMod(wd, pkg, path, method, handler string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
		Pkg:        pkg,
		Path:       path,
		Method:     method,
		Handler:    handler,
	}
}

type CodeMod struct {
	WorkingDir string
	Pkg        string
	Path       string
	Method     string
	Handler    string
}

func (cm *CodeMod) Name() string {
	return "route"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	err := codemod.AddImports(path.Join(cm.WorkingDir, "pkg", cm.Pkg, "router.go"), []string{
		"net/http",
	})
	if err != nil {
		return cerrors.New(err, "failed to add imports to router.go", nil)
	}

	err = codemod.InsertRoute(path.Join(cm.WorkingDir, "pkg", cm.Pkg, "router.go"), codemod.InsertRouteParams{
		Path:        cm.Path,
		Method:      cm.Method,
		HandlerName: cm.Handler,
		HandlerBody: "",
	})
	if err != nil {
		return cerrors.New(err, "failed to insert route", nil)
	}

	return nil
}
