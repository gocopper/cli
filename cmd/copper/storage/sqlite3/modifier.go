package sqlite3

import (
	"path"

	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var (
		csqlConfig = `
		
[csql]
dialect="sqlite3"
dsn="./{{.DB}}.db"
`
	)

	return codemod.
		OpenDir(wd).
		ExtractData(codemod.ExtractGoModulePath()).
		ModifyData(func(data map[string]string) {
			data["DB"] = path.Base(data["Module"])
		}).
		OpenFile("./config/base.toml").
		Apply(codemod.AppendText(csqlConfig)).
		CloseAndOpen("./cmd/app/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/mattn/go-sqlite3\""})).
		CloseAndOpen("./cmd/migrate/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/mattn/go-sqlite3\""})).
		CloseAndDone()
}
