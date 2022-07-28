package postgres

import (
	"github.com/gocopper/cli/pkg/codemod"
)

func Apply(wd string) error {
	var (
		csqlBaseConfig = `

[csql]
dialect="postgres"
`
		csqlDevConfig = `

[csql]
dsn="user=postgres password=1234 host=127.0.0.1 port=5432 dbname=pg sslmode=disable"
`
	)

	return codemod.OpenDir(wd).
		OpenFile("./config/base.toml").
		Apply(codemod.AppendText(csqlBaseConfig)).
		CloseAndOpen("./config/dev.toml").
		Apply(codemod.AppendText(csqlDevConfig)).
		CloseAndOpen("./cmd/app/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/lib/pq\""})).
		CloseAndOpen("./cmd/migrate/wire.go").
		Apply(codemod.AddGoImports([]string{"_ \"github.com/lib/pq\""})).
		CloseAndDone()
}
