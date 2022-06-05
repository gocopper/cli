package tailwind

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

func (cm *CodeMod) Name() string {
	return "tailwind"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	npmInstallCmd := exec.CommandContext(ctx, "npm", "install", "-D", "tailwindcss")
	npmInstallCmd.Dir = path.Join(cm.WorkingDir, "web")

	err := npmInstallCmd.Run()
	if err != nil {
		return cerrors.New(err, "failed to npm install tailwindcss", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = codemod.AddJSONSection(path.Join(cm.WorkingDir, "web", "package.json"), "scripts", map[string]string{
		"build": "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --minify",
		"dev":   "npx tailwindcss -i ./src/styles.css -o ./public/styles.css --watch",
	})
	if err != nil {
		return cerrors.New(err, "failed to add npm scripts", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	_, err = codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, nil, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	err = os.Remove(path.Join(cm.WorkingDir, "web", "public", "styles.css"))
	if err != nil {
		return cerrors.New(err, "failed to remove web/public/styles.css", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
