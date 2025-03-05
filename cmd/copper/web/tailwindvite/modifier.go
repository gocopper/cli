package tailwindvite

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed reacttmpl/*
var reactTemplatesFS embed.FS

//go:embed inertiatmpl/*
var inertiaTemplatesFS embed.FS

func Apply(wd string, inertia bool) error {
	templatesFS := reactTemplatesFS
	if inertia {
		templatesFS = inertiaTemplatesFS
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
