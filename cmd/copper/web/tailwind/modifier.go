package tailwind

import (
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	return codemod.
		New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		ModifyData(func(data map[string]string) {
			data["PackageJSONName"] = path.Base(data["GoModule"])
		}).
		CreateTemplateFiles(templatesFS, nil, true).
		Cd("./web/").
		Remove("./public/styles.css").
		RunCmd("npm", "install", "-D", "tailwindcss").
		OpenFile("./package.json").
		Apply(codemod.ModAddJSONSection("scripts", map[string]string{
			"build": "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --minify",
			"dev":   "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --watch",
		})).
		CloseAndDone()
}
