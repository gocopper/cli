package tailwindvite

import (
	"embed"
	"github.com/gocopper/copper/cerrors"
	"io/fs"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed react/*
var reactTemplatesFS embed.FS

//go:embed inertia/*
var inertiaTemplatesFS embed.FS

func Apply(wd string, inertia bool) error {
	var templatesFS fs.FS

	subDir := "react"
	if inertia {
		subDir = "inertia"
	}

	templatesFS, err := fs.Sub(inertiaTemplatesFS, subDir)
	if err != nil {
		return cerrors.New(err, "failed to open template sub dir", map[string]any{"dir": subDir})
	}

	return codemod.
		OpenDir(wd).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		Cd("./web").
		Apply(codemod.RunCmd("npm", "install", "tailwindcss", "@tailwindcss/vite")).
		OpenFile("vite.config.ts").
		Apply(codemod.InsertLineAtLineNum(3, "import tailwindcss from '@tailwindcss/vite';")).
		Apply(codemod.InsertTextAfter("plugins: [", "tailwindcss(), ")).
		CloseAndDone()
}
