package sourcecode

import (
	"os"
	"path"
)

func ProjectHasWeb(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "web", "src"))

	return err == nil
}

func ProjectHasJS(projectPath string) bool {
	return false
}

func ProjectHasSQL(projectPath string) bool {
	_, err := os.Stat(path.Join(projectPath, "cmd", "migrate"))

	return err == nil
}
