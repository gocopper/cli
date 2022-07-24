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
		New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		ModifyData(func(data map[string]string) {
			data["PackageJSONName"] = path.Base(data["GoModule"])
		}).
		CreateTemplateFiles(templatesFS, nil, true).
		Cd("./web").
		RunCmd("npm", "install", "react", "react-dom").
		RunCmd("npm", "install", "-D", "@types/react", "@types/react-dom", "@vitejs/plugin-react", "vite").
		RenameFile("public/styles.css", "src/styles.css").
		OpenFile("package.json").
		Apply(codemod.ModAddJSONSection("scripts", map[string]string{
			"dev":     "vite",
			"build":   "vite build",
			"preview": "vite preview",
		})).
		Close().
		Cd("../").
		OpenFile("./cmd/app/wire.go").
		Apply(
			codemod.ModAddGoImports([]string{"github.com/gocopper/pkg/vitejs"}),
			codemod.ModAddProviderToWireSet(`vitejs.WireModule`),
		).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.ModAppendText(viteJSConfig)).
		CloseAndDone()
}
