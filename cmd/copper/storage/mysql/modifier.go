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
		OpenDir(wd).
		OpenFile("./config/base.toml").
		Apply(codemod.AppendText(csqlBaseConfig)).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.AppendText(csqlDevConfig)).
		CloseAndOpen("./cmd/app/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/go-sql-driver/mysql\""})).
		CloseAndOpen("./cmd/migrate/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/go-sql-driver/mysql\""})).
		CloseAndDone()
}
