package sourcecode

import (
	"fmt"
	"github.com/gocopper/copper/cerrors"
	"io/ioutil"
	"os"
	"strings"
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

	for i := range deps {
		deps[i] = "\"" + deps[i] + "\""
	}

	return insertText(path, strings.Join(deps, "\n")+"\n", pos+len(importStmt)+1)
}
