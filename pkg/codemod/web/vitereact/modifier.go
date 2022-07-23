package vitereact

import (
	"context"
	"embed"
	"os"
	"os/exec"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

//go:embed tmpl/*
var templatesFS embed.FS

func NewCodeMod(wd, module string) *CodeMod {
	return &CodeMod{
		WorkingDir: wd,
		Module:     module,
	}
}

type CodeMod struct {
	WorkingDir string
	Module     string
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, map[string]string{
		"PackageJSONName": path.Base(cm.Module),
	}, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", map[string]interface{}{
			"wd":     cm.WorkingDir,
			"module": cm.Module,
		})
	}

	err = os.Rename(path.Join(cm.WorkingDir, "web", "public", "styles.css"), path.Join(cm.WorkingDir, "web", "src", "styles.css"))
	if err != nil {
		return cerrors.New(err, "failed to move styles.css", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	installReactCmd := exec.CommandContext(ctx, "npm", "install", "react", "react-dom")
	installReactCmd.Dir = path.Join(cm.WorkingDir, "web")

	err = installReactCmd.Run()
	if err != nil {
		return cerrors.New(err, "failed to npm install react, react-dom", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	installDevReactViteCmd := exec.CommandContext(ctx,
		"npm", "install", "-D", "@types/react", "@types/react-dom", "@vitejs/plugin-react", "vite",
	)
	installDevReactViteCmd.Dir = path.Join(cm.WorkingDir, "web")

	err = installDevReactViteCmd.Run()
	if err != nil {
		return cerrors.New(err, "failed to npm install react types, vite", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddJSONSection(path.Join(cm.WorkingDir, "web", "package.json"), "scripts", map[string]string{
		"dev":     "vite",
		"build":   "vite build",
		"preview": "vite preview",
	})
	if err != nil {
		return cerrors.New(err, "failed to add vite scripts to package.json", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddImports(path.Join(cm.WorkingDir, "pkg", "app", "wire.go"), []string{
		"github.com/gocopper/pkg/vitejs",
	})
	if err != nil {
		return cerrors.New(err, "failed to add vitejs import to pkg/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.InsertWireModuleItems(path.Join(cm.WorkingDir, "pkg", "app", "wire.go"), `
vitejs.WireModule,`)
	if err != nil {
		return cerrors.New(err, "failed to add vitejs.WireModule to pkg/app/wire.go", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AppendTextToFile(path.Join(cm.WorkingDir, "config", "dev.toml"), `

[vitejs]
dev_mode=true
`)
	if err != nil {
		return cerrors.New(err, "failed add vitejs config to dev.toml", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
