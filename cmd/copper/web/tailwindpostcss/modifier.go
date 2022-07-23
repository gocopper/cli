package tailwindpostcss

import (
	"context"
	"embed"
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
	_, err := codemod.CreateTemplateFiles(templatesFS, cm.WorkingDir, nil, true)
	if err != nil {
		return cerrors.New(err, "failed to create template dir", map[string]interface{}{
			"wd":     cm.WorkingDir,
			"module": cm.Module,
		})
	}

	// Assumes package.json is already created
	npmInstallCmd := exec.CommandContext(ctx, "npm", "install", "-D", "tailwindcss", "postcss", "autoprefixer")
	npmInstallCmd.Dir = path.Join(cm.WorkingDir, "web")

	err = npmInstallCmd.Run()
	if err != nil {
		return cerrors.New(err, "failed to npm install tailwindcss, postcss, autoprefixer", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
