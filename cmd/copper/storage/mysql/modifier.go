package mysql

import (
	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var (
		csqlBaseConfig = `

[csql]
dialect="mysql"
`
		csqlDevConfig = `

[csql]
dsn="user:password@/dbname"
`
	)

	return codemod.
		New(wd).
		OpenFile("./config/base.toml").
		Apply(codemod.ModAppendText(csqlBaseConfig)).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.ModAppendText(csqlDevConfig)).
		CloseAndOpen("./cmd/app/wire.go").
		Apply(codemod.ModAddGoImports([]string{"_ \"github.com/go-sql-driver/mysql\""})).
		CloseAndOpen("./cmd/migrate/wire.go").
		Apply(codemod.ModAddGoImports([]string{"_ \"github.com/go-sql-driver/mysql\""})).
		CloseAndDone()
}
