package sql

import (
	"context"
	"path"
	"text/template"

	"github.com/gocopper/cli/v3/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

func NewQueryCodeMod(wd, pkg, model, field string, list bool) *QueryCodeMod {
	return &QueryCodeMod{
		WorkingDir: wd,
		Pkg:        pkg,
		Model:      model,
		Field:      field,
		List:       list,
	}
}

type QueryCodeMod struct {
	WorkingDir string
	Pkg        string
	Model      string
	Field      string
	List       bool
}

func (cm *QueryCodeMod) Name() string {
	return "sql:query"
}

func (cm *QueryCodeMod) Apply(ctx context.Context) error {
	if cm.Field == "" {
		return cm.applyScan(ctx)
	}

	return cm.applyQueryWithField(ctx)
}

func (cm *QueryCodeMod) applyScan(ctx context.Context) error {
	var fp = path.Join(cm.WorkingDir, "pkg", cm.Pkg, "repo.go")

	t := template.Must(template.New("Repo#ScaffoldScan").Parse(`
func (r *Repo) {{.method}}(ctx context.Context) ({{.returnType}}, error) {
	var {{.modelVar}} {{.modelVarType}}

	err := csql.GetConn(ctx, r.db).
		Find(&{{.modelVar}}).
		Error
	if err != nil {
	    return nil, cerrors.New(err, "failed to scan {{.modelVar}}", nil)
    }

	return {{.returnVar}}, nil
}
`))

	err := codemod.AddImports(fp, []string{
		"context",
		"github.com/gocopper/copper/csql",
		"github.com/gocopper/copper/cerrors",
	})
	if err != nil {
		return cerrors.New(err, "failed to add imports to repo.go", nil)
	}

	err = codemod.AppendTemplateToFile(fp, t, map[string]interface{}{
		"method":       "List" + strcase.ToCamel(inflection.Plural(cm.Model)),
		"model":        cm.Model,
		"modelVar":     strcase.ToLowerCamel(inflection.Plural(cm.Model)),
		"modelVarType": "[]" + cm.Model,
		"returnType":   "[]" + cm.Model,
		"returnVar":    strcase.ToLowerCamel(inflection.Plural(cm.Model)),
	})
	if err != nil {
		return cerrors.New(err, "failed to add query method to repo.go", nil)
	}

	return nil
}

func (cm *QueryCodeMod) applyQueryWithField(ctx context.Context) error {
	var fp = path.Join(cm.WorkingDir, "pkg", cm.Pkg, "repo.go")

	t := template.Must(template.New("Repo#ScaffoldQuery").Parse(`
func (r *Repo) {{.method}}(ctx context.Context, {{.fieldVar}} {{.fieldType}}) ({{.returnType}}, error) {
	var {{.modelVar}} {{.modelVarType}}

	err := csql.GetConn(ctx, r.db).
		Where({{.model}}{ {{.field}}: {{.fieldVar}}}).
		{{.gormMethod}}(&{{.modelVar}}).
		Error
	if err != nil {
	    return nil, cerrors.New(err, "failed to query {{.modelVar}}", map[string]interface{}{
	        "{{.fieldVar}}": {{.fieldVar}},
        })
    }

	return {{.returnVar}}, nil
}
`))

	method := "Get" + strcase.ToCamel(cm.Model) + "By" + cm.Field
	if cm.List {
		method = "List" + strcase.ToCamel(inflection.Plural(cm.Model)) + "By" + cm.Field
	}

	modelVar := strcase.ToLowerCamel(cm.Model)
	if cm.List {
		modelVar = strcase.ToLowerCamel(inflection.Plural(cm.Model))
	}

	modelVarType := cm.Model
	if cm.List {
		modelVarType = "[]" + cm.Model
	}

	gormMethod := "First"
	if cm.List {
		gormMethod = "Find"
	}

	returnType := "*" + modelVarType
	if cm.List {
		returnType = "[]" + cm.Model
	}

	returnVar := "&" + modelVar
	if cm.List {
		returnVar = modelVar
	}

	err := codemod.AddImports(fp, []string{
		"context",
		"github.com/gocopper/copper/csql",
		"github.com/gocopper/copper/cerrors",
	})
	if err != nil {
		return cerrors.New(err, "failed to add imports to repo.go", nil)
	}

	err = codemod.AppendTemplateToFile(fp, t, map[string]interface{}{
		"method":       method,
		"field":        cm.Field,
		"fieldVar":     strcase.ToLowerCamel(cm.Field),
		"fieldType":    "string",
		"model":        cm.Model,
		"modelVar":     modelVar,
		"modelVarType": modelVarType,
		"gormMethod":   gormMethod,
		"returnType":   returnType,
		"returnVar":    returnVar,
	})
	if err != nil {
		return cerrors.New(err, "failed to add query method to repo.go", nil)
	}

	return nil
}
