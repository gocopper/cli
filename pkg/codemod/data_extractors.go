package codemod

import (
	"errors"
	"io/ioutil"
	"path"

	"golang.org/x/mod/modfile"
)

type DataExtractorFn func(wd string) (map[string]string, error)

func ExtractGoModulePath() DataExtractorFn {
	return func(wd string) (map[string]string, error) {
		goModData, err := ioutil.ReadFile(path.Join(wd, "go.mod"))
		if err != nil {
			return nil, err
		}

		modPath := modfile.ModulePath(goModData)
		if modPath == "" {
			return nil, errors.New("mod path is empty")
		}

		return map[string]string{
			"GoModule": modPath,
		}, nil
	}

}
