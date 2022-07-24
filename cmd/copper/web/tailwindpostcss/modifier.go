package tailwindpostcss

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	return codemod.
		New(wd).
		CreateTemplateFiles(templatesFS, nil, true).
		Cd("./web").
		RunCmd("npm", "install", "-D", "tailwindcss", "postcss", "autoprefixer").
		Done()
}
