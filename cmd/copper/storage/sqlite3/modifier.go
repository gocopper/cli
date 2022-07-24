package sqlite3

import (
	"fmt"
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var (
		csqlConfig string
	)

	return codemod.
		New(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		Do(func(data map[string]string) error {
			csqlConfig = fmt.Sprintf(`

[csql]
dialect="sqlite3"
dsn="./%s.db"
`, path.Base(data["GoModule"]))
			return nil
		}).
		OpenFile("./config/base.toml").
		Apply(codemod.ModAppendText(csqlConfig)).
		CloseAndOpen("./cmd/app/wire.go").
		Apply(codemod.ModAddGoImports([]string{"_ \"github.com/mattn/go-sqlite3\""})).
		CloseAndOpen("./cmd/migrate/wire.go").
		Apply(codemod.ModAddGoImports([]string{"_ \"github.com/mattn/go-sqlite3\""})).
		CloseAndDone()
}
