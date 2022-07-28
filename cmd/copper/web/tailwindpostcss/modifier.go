package tailwindpostcss

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
		Apply(codemod.RunCmd("npm", "install", "-D", "tailwindcss", "postcss", "autoprefixer")).
		Done()
}
