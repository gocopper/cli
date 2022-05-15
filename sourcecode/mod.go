package sourcecode

import (
	"io/ioutil"
	"path"

	"github.com/gocopper/copper/cerrors"
	"golang.org/x/mod/modfile"
)

func GetGoModulePath(projectDir string) (string, error) {
	goModData, err := ioutil.ReadFile(path.Join(projectDir, "go.mod"))
	if err != nil {
		return "", cerrors.New(err, "failed to read go.mod", map[string]interface{}{
			"projectDir": projectDir,
		})
	}

	modPath := modfile.ModulePath(goModData)
	if modPath == "" {
		return "", cerrors.New(nil, "mod path is empty", map[string]interface{}{
			"projectDir": projectDir,
		})
	}

	return modPath, nil
}
