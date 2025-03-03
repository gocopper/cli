package tailwindvite

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
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
