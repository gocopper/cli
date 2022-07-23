package frontendnone

import (
	"context"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

func NewCodeMod(wd string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
	}
}

type CodeMod struct {
	WorkingDir string
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	err := codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "cmd", "app", "wire.go"), `
chttp.WireModuleEmptyHTML,`)
	if err != nil {
		return cerrors.New(err, "failed to add web wire modules to cmd/app/wire.go", nil)
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), []string{
		"net/http",
	})
	if err != nil {
		return cerrors.New(err, "failed to add net/http import to pkg/app/router.go", nil)
	}

	err = codemod.InsertRoute(path.Join(cm.WorkingDir, "pkg", "app", "router.go"), codemod.InsertRouteParams{
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
	})
	if err != nil {
		return cerrors.New(err, "failed to insert route for index route", nil)
	}

	return nil
}
