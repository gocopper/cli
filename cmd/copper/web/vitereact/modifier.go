package vitereact

import (
	"embed"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

//go:embed tmpl/*
var templatesFS embed.FS

func Apply(wd string) error {
	var (
		viteJSConfig = `

[vitejs]
dev_mode=true
`
	)

	return codemod.
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		ModifyData(func(data map[string]string) {
			data["PackageJSONName"] = path.Base(data["GoModule"])
		}).
		Apply(codemod.CreateTemplateFiles(templatesFS, nil, true)).
		Cd("./web").
		Apply(
			codemod.RunCmd("npm", "install", "react", "react-dom"),
			codemod.RunCmd("npm", "install", "-D", "@types/react", "@types/react-dom", "@vitejs/plugin-react", "vite"),
			codemod.RenameFile("public/styles.css", "src/styles.css"),
		).
		OpenFile("package.json").
		Apply(codemod.AddJSONSection("scripts", map[string]string{
			"dev":     "vite",
			"build":   "vite build",
			"preview": "vite preview",
		})).
		Close().
		Cd("../").
		OpenFile("./cmd/app/wire.go").
		Apply(
			codemod.AddGoImports([]string{"github.com/gocopper/pkg/vitejs"}),
			codemod.AddProviderToWireSet(`vitejs.WireModule`),
		).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.AppendText(viteJSConfig)).
		CloseAndDone()
}
