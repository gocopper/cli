package cli

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gocopper/cli/sourcecode"
	"github.com/gocopper/copper/cerrors"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

const (
	ScaffolderRepo        = "Repository"
	ScaffolderRepoQuery   = "Repository\t-> Query"
	ScaffolderRepoScan    = "Repository\t-> Scan"
	ScaffolderRepoSave    = "Repository\t-> Save"
	ScaffolderRouter      = "Router"
	ScaffolderRouterRoute = "Router\t-> Route"
)

//go:embed tmpl
var TemplatesFS embed.FS

type Scaffold struct {
	initBaseTmpl  *template.Template
	initSQLTmpl   *template.Template
	initWebTmpl   *template.Template
	initWebJSTmpl *template.Template

	pkgTmpl *template.Template
	term    *Terminal
	make    *Make
}

func NewScaffold() *Scaffold {
	initBaseTmpl, err := template.ParseFS(TemplatesFS, "tmpl/init/base/*.tmpl")
	if err != nil {
		panic(err)
	}

	initSQLTmpl, err := template.ParseFS(TemplatesFS, "tmpl/init/sql/*.tmpl")
	if err != nil {
		panic(err)
	}

	initWebTmpl, err := template.ParseFS(TemplatesFS, "tmpl/init/web/*.tmpl")
	if err != nil {
		panic(err)
	}

	initWebJSTmpl, err := template.ParseFS(TemplatesFS, "tmpl/init/webjs/*.tmpl")
	if err != nil {
		panic(err)
	}

	pkgTmpl, err := template.ParseFS(TemplatesFS, "tmpl/*.tmpl")
	if err != nil {
		panic(err)
	}

	return &Scaffold{
		initBaseTmpl:  initBaseTmpl,
		initSQLTmpl:   initSQLTmpl,
		initWebTmpl:   initWebTmpl,
		initWebJSTmpl: initWebJSTmpl,
		pkgTmpl:       pkgTmpl,
		make:          NewMake(),
		term:          NewTerminal(),
	}
}

func (s *Scaffold) Init() bool {
	var (
		module string
		web    bool
		js     bool
		sql    bool
	)

	err := survey.AskOne(&survey.Input{
		Message: "What's the module name for your project?",
		Help:    "ex. github.com/myuser/myproject",
	}, &module)
	if err != nil {
		return false
	}

	err = survey.AskOne(&survey.Confirm{
		Message: "Will your app use SQL for storage?",
	}, &sql)
	if err != nil {
		return false
	}

	err = survey.AskOne(&survey.Confirm{
		Message: "Will your app serve HTML pages?",
	}, &web)
	if err != nil {
		return false
	}

	/*
		Disable special handling for JS or CSS frameworks
		if web {
			err = survey.AskOne(&survey.Confirm{
				Message: "Will you be using JS or CSS frameworks? (React, Vue, Tailwind, Boostrap, etc.)",
			}, &js)
			if err != nil {
				return false
			}
		}
	*/

	project := strcase.ToLowerCamel(path.Base(module))

	workDir, err := filepath.Abs(path.Base(module))
	if err != nil {
		return false
	}

	params := map[string]string{
		"module":      module,
		"project":     project,
		"initProject": "Init" + strcase.ToCamel(path.Base(module)),
	}

	s.term.Section("Create Project Files")

	if ok := s.scaffoldTemplate(s.initBaseTmpl, workDir, params); !ok {
		return false
	}

	if web {
		err = os.Remove(path.Join(workDir, "pkg", "app", "handler.go"))
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add delete pkg/app/handler.go", nil))
			return false
		}

		err = os.Remove(path.Join(workDir, "config", "base.toml"))
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add delete config/base.toml", nil))
			return false
		}

		err = os.Remove(path.Join(workDir, "config", "dev.toml"))
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add delete config/dev.toml", nil))
			return false
		}

		tmpl := s.initWebTmpl
		if js {
			tmpl = s.initWebJSTmpl
		}

		ok := s.scaffoldTemplate(tmpl, workDir, params)
		if !ok {
			return false
		}

		err = sourcecode.AddImports(path.Join(workDir, "cmd", "app", "wire.go"), []string{
			path.Join(module, "web"),
			path.Join(module, "web/build"),
		})
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add deps to pkg/app/wire.go", nil))
			return false
		}

		err = sourcecode.InsertWireModuleItem(path.Join(workDir, "cmd", "app"), `
app.NewRouter,
wire.Struct(new(app.NewRouterParams), "*"),

wire.InterfaceValue(new(chttp.HTMLDir), web.HTMLDir),
wire.InterfaceValue(new(chttp.StaticDir), build.StaticDir),
web.HTMLRenderFuncs,
`)
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add web module to app", nil))
			return false
		}
	}

	if sql {
		t := template.Must(template.New("csql#Config").Parse(`
[csql]
dialect="sqlite"
dsn="./{{.project}}.db"
`))

		err = sourcecode.AppendTemplateToFile(t, map[string]interface{}{
			"project": project,
		}, path.Join(workDir, "config", "base.toml"))
		if err != nil {
			s.term.TaskFailed(cerrors.New(err, "Failed to update config/base.toml", nil))
			return false
		}

		err = sourcecode.AddImports(path.Join(workDir, "cmd", "app", "wire.go"), []string{
			"github.com/gocopper/copper/csql",
		})
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add csql deps to pkg/app/wire.go", nil))
			return false
		}

		err = sourcecode.InsertWireModuleItem(path.Join(workDir, "cmd", "app"), `
csql.WireModule,
`)
		if err != nil {
			s.term.Error(cerrors.New(err, "Failed to add csql module to app", nil))
			return false
		}

		ok := s.scaffoldTemplate(s.initSQLTmpl, workDir, params)
		if !ok {
			return false
		}
	}

	s.term.Section("First Commands")
	s.term.Box(fmt.Sprintf(`cd %s
copper build
copper watch`, project))

	return true
}

func (s *Scaffold) scaffoldTemplate(tmpl *template.Template, dir string, params map[string]string) bool {
	for _, t := range tmpl.Templates() {
		filePath := strings.ReplaceAll(t.Name(), "$", "/")
		filePath = strings.Replace(filePath, ".tmpl", "", 1)
		filePath = strings.Replace(filePath, "^", ".", 1)
		filePath = path.Join(dir, filePath)

		s.term.InProgressTask(fmt.Sprintf("Create %s", filePath))

		err := os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			s.term.TaskFailed(cerrors.New(err, fmt.Sprintf("Failed to create directories for %s", filePath), nil))
			return false
		}

		err = sourcecode.CreateTemplateFile(filePath, t, params)
		if err != nil {
			s.term.TaskFailed(cerrors.New(err, fmt.Sprintf("Failed to create %s", filePath), nil))
			return false
		}

		s.term.TaskSucceeded()
	}

	return true
}

func (s *Scaffold) Run(ctx context.Context, pkg string) bool {
	ok := s.scaffoldPkgIfNeeded(pkg)
	if !ok {
		return false
	}

	scaffolder, ok := s.promptForScaffolder(pkg)
	if !ok {
		return false
	}

	switch scaffolder {
	case ScaffolderRepo:
		return s.scaffoldRepository(pkg)
	case ScaffolderRepoSave:
		return s.scaffoldRepositorySave(pkg)
	case ScaffolderRepoQuery:
		return s.scaffoldRepositoryQuery(pkg)
	case ScaffolderRepoScan:
		return s.scaffoldRepositoryScan(pkg)
	case ScaffolderRouter:
		return s.scaffoldRouter(pkg)
	case ScaffolderRouterRoute:
		return s.scaffoldRouterRoute(pkg)
	}

	return true
}

func (s *Scaffold) promptForScaffolder(pkg string) (string, bool) {
	var (
		repoFilePath   = path.Join("pkg", pkg, "repo.go")
		routerFilePath = path.Join("pkg", pkg, "router.go")
		choice         string
	)

	opts := make([]string, 0)

	if sourcecode.ProjectHasSQL(".") {
		_, err := os.Stat(repoFilePath)
		if err == nil {
			opts = append(opts, ScaffolderRepoQuery, ScaffolderRepoScan, ScaffolderRepoSave)
		} else if sourcecode.ProjectHasSQL(".") {
			opts = append(opts, ScaffolderRepo)
		}
	}

	_, err := os.Stat(routerFilePath)
	if err == nil {
		opts = append(opts, ScaffolderRouterRoute)
	} else {
		opts = append(opts, ScaffolderRouter)
	}

	err = survey.AskOne(&survey.Select{
		Message: "Choose a scaffolder",
		Options: opts,
	}, &choice)
	if err != nil {
		return "", false
	}

	return choice, true
}

func (s *Scaffold) promptForModel(pkg string) (sourcecode.Struct, bool) {
	var model string

	modelStructs, err := sourcecode.FindStructs(path.Join("pkg", pkg, "models.go"))
	if err != nil {
		s.term.Error(cerrors.New(err, "Failed to find model structs in models.go", nil))
		return sourcecode.Struct{}, false
	}

	if len(modelStructs) == 0 {
		s.term.Error(cerrors.New(nil, "No model found in models.go", nil))
		return sourcecode.Struct{}, false
	}

	modelNames := make([]string, len(modelStructs))
	for i, s := range modelStructs {
		modelNames[i] = s.Name
	}

	err = survey.AskOne(&survey.Select{
		Message: "Choose a model",
		Options: modelNames,
	}, &model)
	if err != nil {
		return sourcecode.Struct{}, false
	}

	for _, s := range modelStructs {
		if s.Name == model {
			return s, true
		}
	}

	return sourcecode.Struct{}, false
}

func (s *Scaffold) promptForModelField(pkg string) (sourcecode.Struct, sourcecode.StructField, bool) {
	var chosenFieldName string

	model, ok := s.promptForModel(pkg)
	if !ok {
		return sourcecode.Struct{}, sourcecode.StructField{}, false
	}

	modelFields := make([]string, 0)
	for _, f := range model.Fields {
		modelFields = append(modelFields, f.Name)
	}

	err := survey.AskOne(&survey.Select{
		Message: "Choose field to query on",
		Options: modelFields,
	}, &chosenFieldName)
	if err != nil {
		return sourcecode.Struct{}, sourcecode.StructField{}, false
	}

	for _, f := range model.Fields {
		if f.Name == chosenFieldName {
			return model, f, true
		}
	}

	return sourcecode.Struct{}, sourcecode.StructField{}, false
}

func (s *Scaffold) scaffoldRouter(pkg string) bool {
	s.term.InProgressTask("Create router.go")
	err := sourcecode.CreateTemplateFile(path.Join("pkg", pkg, "router.go"), s.pkgTmpl.Lookup("router.go.tmpl"), map[string]string{
		"pkg": pkg,
	})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create router.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Update wire.go")
	err = sourcecode.InsertWireModuleItem(path.Join("pkg", pkg), `
	wire.Struct(new(NewRouterParams), "*"),
	NewRouter,
`)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update wire.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return true
}

func (s *Scaffold) scaffoldRouterRoute(pkg string) bool {
	var (
		urlPath       string
		handlerMethod string
		httpMethod    string

		routeSb strings.Builder
	)

	err := survey.AskOne(&survey.Input{
		Message: "What is the URL path?",
		Help:    "ex. /api/profiles/{id}",
	}, &urlPath)
	if err != nil {
		return false
	}

	err = survey.AskOne(&survey.Input{
		Message: "What is the route handler called?",
		Help:    "ex. GetUserProfile, UserPreferencesPage",
	}, &handlerMethod)
	if err != nil {
		return false
	}

	err = survey.AskOne(&survey.Select{
		Message: "Select HTTP method",
		Options: []string{"Get", "Post", "Put", "Patch", "Delete"},
	}, &httpMethod)
	if err != nil {
		return false
	}

	routeT := template.Must(template.New("Router#ScaffoldRoute.route").Parse(`{
	Path:        "{{.path}}",
	Methods:     []string{http.Method{{.httpMethod}}},
	Handler:     ro.{{.handlerMethod}},
},
`))

	s.term.InProgressTask("Add new route to router.go")

	err = routeT.Execute(&routeSb, map[string]string{
		"path":          urlPath,
		"httpMethod":    httpMethod,
		"handlerMethod": handlerMethod,
	})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to add new route to router.go", nil))
		return false
	}

	routerFilePath := path.Join("pkg", pkg, "router.go")

	offset, err := sourcecode.PosToInsertNewRouteDecl(routerFilePath)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to find position to insert new route in router.go", nil))
		return false
	}

	err = sourcecode.InsertTextToFile(routerFilePath, routeSb.String(), offset)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to insert route router.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	handlerT := template.Must(template.New("Router#ScaffoldRoute.handler").Parse(`
func (ro *Router) {{.handlerMethod}}(w http.ResponseWriter, r *http.Request) {
}
`))

	s.term.InProgressTask("Add route handler to router.go")
	err = sourcecode.AppendTemplateToFile(handlerT, map[string]interface{}{
		"handlerMethod": handlerMethod,
	}, routerFilePath)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to add route handler to router.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return true
}

func (s *Scaffold) scaffoldRepositoryScan(pkg string) bool {
	model, ok := s.promptForModel(pkg)
	if !ok {
		return false
	}

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

	s.term.InProgressTask("Update repo.go")
	err := sourcecode.AppendTemplateToFile(t, map[string]interface{}{
		"method":       "List" + strcase.ToCamel(inflection.Plural(model.Name)),
		"model":        model.Name,
		"modelVar":     strcase.ToLowerCamel(inflection.Plural(model.Name)),
		"modelVarType": "[]" + model.Name,
		"returnType":   "[]" + model.Name,
		"returnVar":    strcase.ToLowerCamel(inflection.Plural(model.Name)),
	}, path.Join("pkg", pkg, "repo.go"))
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update repo.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return false
}

func (s *Scaffold) scaffoldRepositoryQuery(pkg string) bool {
	var listQuery bool

	model, field, ok := s.promptForModelField(pkg)
	if !ok {
		return false
	}

	err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Query returns a list of %s?", model.Name),
		Default: false,
	}, &listQuery)
	if err != nil {
		return false
	}

	method := "Get" + strcase.ToCamel(model.Name) + "By" + field.Name
	if listQuery {
		method = "List" + strcase.ToCamel(inflection.Plural(model.Name)) + "By" + field.Name
	}

	modelVar := strcase.ToLowerCamel(model.Name)
	if listQuery {
		modelVar = strcase.ToLowerCamel(inflection.Plural(model.Name))
	}

	modelVarType := model.Name
	if listQuery {
		modelVarType = "[]" + model.Name
	}

	gormMethod := "First"
	if listQuery {
		gormMethod = "Find"
	}

	returnType := "*" + modelVarType
	if listQuery {
		returnType = "[]" + model.Name
	}

	returnVar := "&" + modelVar
	if listQuery {
		returnVar = modelVar
	}

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

	s.term.InProgressTask("Update repo.go")
	err = sourcecode.AppendTemplateToFile(t, map[string]interface{}{
		"method":       method,
		"field":        field.Name,
		"fieldVar":     strcase.ToLowerCamel(field.Name),
		"fieldType":    "string",
		"model":        model.Name,
		"modelVar":     modelVar,
		"modelVarType": modelVarType,
		"gormMethod":   gormMethod,
		"returnType":   returnType,
		"returnVar":    returnVar,
	}, path.Join("pkg", pkg, "repo.go"))
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update repo.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return false
}

func (s *Scaffold) scaffoldRepositorySave(pkg string) bool {
	model, ok := s.promptForModel(pkg)
	if !ok {
		return false
	}

	t := template.Must(template.New("Repo#ScaffoldModelSave").Parse(`
func (r *Repo) Save{{.model}}(ctx context.Context, {{.modelVar}} *{{.model}}) error {
	return csql.GetConn(ctx, r.db).Save({{.modelVar}}).Error
}
`))

	s.term.InProgressTask("Update repo.go")
	err := sourcecode.AppendTemplateToFile(t, map[string]string{
		"model":    model.Name,
		"modelVar": strcase.ToLowerCamel(model.Name),
	}, path.Join("pkg", pkg, "repo.go"))
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update repo.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return true
}

func (s *Scaffold) scaffoldRepository(pkg string) bool {
	const (
		repoTmplName      = "repo.go.tmpl"
		migrationTmplName = "migration.go.tmpl"
	)

	var (
		repoFilePath      = path.Join("pkg", pkg, "repo.go")
		migrationFilePath = path.Join("pkg", pkg, "migration.go")
		repoTmpl          = s.pkgTmpl.Lookup(repoTmplName)
		migrationTmpl     = s.pkgTmpl.Lookup(migrationTmplName)
	)

	s.term.InProgressTask("Create repo.go")
	err := sourcecode.CreateTemplateFile(repoFilePath, repoTmpl, map[string]string{"pkg": pkg})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create repo.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Create migration.go")
	err = sourcecode.CreateTemplateFile(migrationFilePath, migrationTmpl, map[string]string{"pkg": pkg})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create migration.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Update wire.go")
	err = sourcecode.InsertWireModuleItem(path.Join("pkg", pkg), `
	NewRepo,
	NewMigration,
`)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update wire.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	return true
}

func (s *Scaffold) scaffoldPkgIfNeeded(name string) bool {
	pkgDir := path.Join("pkg", name)

	_, err := os.Stat(pkgDir)
	if err == nil {
		return true
	}

	scaffoldPkg := false

	err = survey.AskOne(&survey.Confirm{
		Message: "Package does not exist. Scaffold a new one?",
		Default: false,
	}, &scaffoldPkg)
	if err != nil || !scaffoldPkg {
		return false
	}

	s.term.Section("Scaffold Package")

	projectModule, err := sourcecode.GetGoModulePath(".")
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to get project module", nil))
		return false
	}

	pkgImportPath := path.Join(projectModule, "pkg", name)

	s.term.InProgressTask("Create pkg dir")
	err = os.Mkdir(pkgDir, 0755)
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create pkg dir", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Create models.go")
	err = sourcecode.CreateTemplateFile(path.Join(pkgDir, "models.go"), s.pkgTmpl.Lookup("models.go.tmpl"), map[string]string{
		"pkg": name,
	})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create models.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Create wire.go")
	err = sourcecode.CreateTemplateFile(path.Join(pkgDir, "wire.go"), s.pkgTmpl.Lookup("wire.go.tmpl"), map[string]string{
		"pkg": name,
	})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to create wire.go", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.InProgressTask("Update pkg/app/wire.go")
	err = sourcecode.AddImports(path.Join("pkg", "app", "wire.go"), []string{pkgImportPath})
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update pkg/app/wire.go imports", nil))
		return false
	}

	err = sourcecode.InsertWireModuleItem(path.Join("pkg", "app"), fmt.Sprintf(`
%s.WireModule,
`, name))
	if err != nil {
		s.term.TaskFailed(cerrors.New(err, "Failed to update pkg/app/wire.go WireModule", nil))
		return false
	}
	s.term.TaskSucceeded()

	s.term.LineBreak()

	return true
}
