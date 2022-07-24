package appbase

import (
	"embed"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd, module string) error {
	return codemod.
		New(wd).
		CreateTemplateFiles(templatesFS, map[string]string{
			"GoModule": module,
		}, true).
		Done()
}
