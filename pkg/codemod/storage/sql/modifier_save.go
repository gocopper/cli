package sql

import (
	"context"
	"path"
	"text/template"

	"github.com/gocopper/cli/v3/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
	"github.com/iancoleman/strcase"
)

func NewSaveCodeMod(wd, pkg, model string) *SaveCodeMod {
	return &SaveCodeMod{
		WorkingDir: wd,
		Pkg:        pkg,
		Model:      model,
	}
}

type SaveCodeMod struct {
	WorkingDir string
	Pkg        string
	Model      string
}

func (cm *SaveCodeMod) Name() string {
	return "sql:save"
}

func (cm *SaveCodeMod) Apply(ctx context.Context) error {
	var fp = path.Join(cm.WorkingDir, "pkg", cm.Pkg, "repo.go")

	t := template.Must(template.New("Repo#ScaffoldModelSave").Parse(`
func (r *Repo) Save{{.model}}(ctx context.Context, {{.modelVar}} *{{.model}}) error {
	return csql.GetConn(ctx, r.db).Save({{.modelVar}}).Error
}
`))

	err := codemod.AddImports(fp, []string{
		"context",
		"github.com/gocopper/copper/csql",
	})
	if err != nil {
		return cerrors.New(err, "failed to add imports to repo.go", nil)
	}

	err = codemod.AppendTemplateToFile(fp, t, map[string]string{
		"model":    cm.Model,
		"modelVar": strcase.ToLowerCamel(cm.Model),
	})
	if err != nil {
		return cerrors.New(err, "failed to add Save method to repo.go", nil)
	}

	return nil
}
