package sourcecode

import (
	"os"
	"path"
)

func ProjectHasWeb(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "pkg", "web", "vite.config.js"))

	return err == nil
}

func ProjectHasSQL(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "cmd", "migrate"))

	return err == nil
}