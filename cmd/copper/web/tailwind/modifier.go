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
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		ModifyData(func(data map[string]string) {
			data["PackageJSONName"] = path.Base(data["GoModule"])
		}).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		Cd("./web/").
		Apply(
			codemod.RemoveFile("./public/styles.css"),
			codemod.RunCmd("npm", "install", "-D", "tailwindcss"),
		).
		OpenFile("./package.json").
		Apply(codemod.AddJSONSection("scripts", map[string]string{
			"build": "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --minify",
			"dev":   "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --watch",
		})).
		CloseAndDone()
}
