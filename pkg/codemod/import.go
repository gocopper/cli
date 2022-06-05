package codemod

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gocopper/copper/cerrors"
)

func AddImports(path string, deps []string) error {
	const importStmt = "import ("

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file; %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file; %v", err)
	}

	pos := strings.Index(string(data), importStmt)
	if pos == -1 {
		return cerrors.New(err, "failed to find existing 'import' statement in file", nil)
	}

	depsWithQuotes := make([]string, 0, len(deps))
	for i := range deps {
		d := "\"" + deps[i] + "\""

		if strings.Contains(string(data), d) {
			continue
		}

		depsWithQuotes = append(depsWithQuotes, d)
	}

	return InsertTextToFile(path, strings.Join(depsWithQuotes, "\n")+"\n", pos+len(importStmt)+1)
}
