package tailwindpostcss

import (
	"context"
	"os/exec"
	"path"

	"github.com/gocopper/copper/cerrors"
)

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
	return "tailwind:postcss"
}

func (cm *CodeMod) Apply(ctx context.Context) error {
	// Assumes package.json is already created

	npmInstallCmd := exec.CommandContext(ctx, "npm", "install", "-D", "tailwindcss", "postcss", "autoprefixer")
	npmInstallCmd.Dir = path.Join(cm.WorkingDir, "web")

	err := npmInstallCmd.Run()
	if err != nil {
		return cerrors.New(err, "failed to npm install tailwindcss, postcss, autoprefixer", map[string]interface{}{
			"wd": cm.WorkingDir,
		})
	}

	return nil
}
