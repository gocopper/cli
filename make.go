package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gocopper/cli/sourcecode"

	"github.com/gocopper/copper/cerrors"
	"github.com/otiai10/copy"
	"github.com/radovskyb/watcher"
)

type Make struct {
	term    *Terminal
	webTmpl *template.Template
}

func NewMake() *Make {
	webTmpl, err := template.ParseFS(TemplatesFS, "tmpl/web/*.tmpl")
	if err != nil {
		panic(err)
	}

	return &Make{
		term:    NewTerminal(),
		webTmpl: webTmpl,
	}
}

type WatchParams struct {
	ProjectPath string
}

func (m *Make) Watch(ctx context.Context, p WatchParams) bool {
	ok := m.Run(ctx, RunParams{
		ProjectPath: p.ProjectPath,
		JS:          sourcecode.ProjectHasJS(p.ProjectPath),
	})
	if !ok {
		return false
	}

	w := watcher.New()

	w.SetMaxEvents(1)

	err := w.AddRecursive(path.Join(p.ProjectPath, "pkg"))
	if err != nil {
		m.term.Error(cerrors.New(err, "Failed to watch pkg directory", nil))
		return false
	}

	r := regexp.MustCompile(".*.go$")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	runCtx, cancelRun := context.WithCancel(ctx)

	go func() {
		err := w.Start(time.Millisecond * 500)
		if err != nil {
			m.term.Error(cerrors.New(err, "Failed to start watching pkg directory", nil))
		}
	}()

	for {
		select {
		case <-w.Event:
			cancelRun()
			runCtx.Done()

			runCtx, cancelRun = context.WithCancel(ctx)

			go func() {
				_ = m.Run(runCtx, RunParams{
					ProjectPath: p.ProjectPath,
					App:         true,
				})
			}()
		case err := <-w.Error:
			cancelRun()
			m.term.Error(cerrors.New(err, "Error while watching pkg directory", nil))
			return false
		case <-w.Closed:
			cancelRun()
			return true
		}
	}
}

type RunParams struct {
	ProjectPath string
	App         bool
	JS          bool
}

func (m *Make) Run(ctx context.Context, p RunParams) bool {
	var (
		binary = path.Join(p.ProjectPath, "build", path.Base(p.ProjectPath)+"-"+"app.out")
		cmd    = exec.CommandContext(ctx, binary)
	)

	p.JS = p.JS && sourcecode.ProjectHasJS(p.ProjectPath)

	ok := m.Build(ctx, BuildParams{
		ProjectPath: p.ProjectPath,
		App:         p.App,
		JS:          p.JS,
	})
	if !ok {
		return false
	}

	if p.JS {
		m.term.Section("Start server(s)")

		m.term.InProgressTask("Start vite")
		err := npmRunDev(ctx, p.ProjectPath)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to start vite", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if p.App {
		m.term.Section("App Logs")
		err := cmd.Run()
		if err != nil {
			return false
		}
	}

	return true
}

type MigrateParams struct {
	ProjectPath string
}

func (m *Make) Migrate(ctx context.Context, p MigrateParams) bool {
	var (
		binary = path.Join(p.ProjectPath, "build", path.Base(p.ProjectPath)+"-"+"migrate.out")
		cmd    = exec.CommandContext(ctx, binary)
	)

	ok := m.Build(ctx, BuildParams{
		ProjectPath: p.ProjectPath,
		Migrate:     true,
	})
	if !ok {
		return false
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	m.term.Section("Run migrations")
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

type BuildParams struct {
	ProjectPath string
	App         bool
	Migrate     bool
	JS          bool
}

func (m *Make) Build(ctx context.Context, p BuildParams) bool {
	p.Migrate = p.Migrate && sourcecode.ProjectHasSQL(p.ProjectPath)
	p.JS = p.JS && sourcecode.ProjectHasJS(p.ProjectPath)

	if sourcecode.ProjectHasWeb(p.ProjectPath) {
		m.term.Section("Embed web files")
	}

	// if there are web files but no vite configured, we have to move web/public dir -> web/build/static here
	if sourcecode.ProjectHasWeb(p.ProjectPath) && !p.JS {
		m.term.InProgressTask("Copy web static assets to web/build/static")
		err := copy.Copy(filepath.Join(p.ProjectPath, "web", "public"), filepath.Join(p.ProjectPath, "web", "build", "static"))
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to copy web static assets", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	if sourcecode.ProjectHasWeb(p.ProjectPath) {
		m.term.InProgressTask("Generate web/build/embed.go")
		err := sourcecode.CreateTemplateFile(path.Join(p.ProjectPath, "web", "build", "embed.go"), m.webTmpl.Lookup("embed.go.tmpl"), nil)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to generate web/build/embed.go", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	m.term.Section("Update Dependencies")

	if p.App || p.Migrate {
		m.term.InProgressTask("go mod tidy")
		err := goModTidy(ctx, p)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to run go mod tidy", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	if p.JS {
		m.term.InProgressTask("npm install")
		err := npmInstall(ctx, p.ProjectPath)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to run npm install", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	if p.App || p.Migrate {
		m.term.Section("Generate Wire files")
	}

	if p.Migrate {
		m.term.InProgressTask("wire cmd/migrate")
		err := wireGen(ctx, p.ProjectPath, path.Join(p.ProjectPath, "cmd", "migrate"))
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to generate wire files in cmd/migrate", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	if p.App {
		m.term.InProgressTask("wire cmd/app")
		err := wireGen(ctx, p.ProjectPath, path.Join(p.ProjectPath, "cmd", "app"))
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to generate wire files in cmd/app", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	m.term.Section("Build bundles & binaries")

	if p.JS {
		m.term.InProgressTask("JS bundle")
		err := npmScript(ctx, p.ProjectPath, "build")
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to bundle JS", nil))
			return false
		}

		m.term.TaskSucceeded()
	}

	if p.Migrate {
		m.term.InProgressTask("Migrate binary")
		err := goBuildMigrate(ctx, p)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to build migrate binary", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	if p.App {
		m.term.InProgressTask("App binary")
		err := goBuildApp(ctx, p)
		if err != nil {
			m.term.TaskFailed(cerrors.New(err, "Failed to build app binary", nil))
			return false
		}
		m.term.TaskSucceeded()
	}

	return true
}

func goBuildApp(ctx context.Context, p BuildParams) error {
	return goBuild(ctx, p, path.Join(p.ProjectPath, "cmd", "app"))
}

func goBuildMigrate(ctx context.Context, p BuildParams) error {
	return goBuild(ctx, p, path.Join(p.ProjectPath, "cmd", "migrate"))
}

func goBuild(ctx context.Context, p BuildParams, main string) error {
	var (
		binary = path.Base(p.ProjectPath) + "-" + path.Base(main) + ".out"
		out    = path.Join(p.ProjectPath, "build", binary)

		// todo: remove csql_sqlite build tag
		cmd = exec.CommandContext(ctx, "go", "build", "-tags", "csql_sqlite", "-o", out, main)
	)

	cmd.Dir = p.ProjectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func npmInstall(ctx context.Context, projectPath string) error {
	var (
		dir = path.Join(projectPath, "web")
		cmd = exec.CommandContext(ctx, "npm", "install")
	)

	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func npmRunDev(ctx context.Context, projectPath string) error {
	var (
		dir = path.Join(projectPath, "web")
		cmd = exec.CommandContext(ctx, "npm", "run", "dev")
		out strings.Builder
	)

	cmd.Dir = dir

	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	go func() {
		_ = cmd.Run()
	}()

	time.Sleep(500 * time.Millisecond)

	if !strings.Contains(out.String(), "server running") {
		return errors.New(out.String())
	}

	return nil
}

func npmScript(ctx context.Context, projectPath, script string) error {
	var (
		dir = path.Join(projectPath, "web")
		cmd = exec.CommandContext(ctx, "npm", "run", script)
	)

	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func wireGen(ctx context.Context, projectPath, main string) error {
	cmd := exec.CommandContext(ctx, "wire", "gen", main)

	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func goModTidy(ctx context.Context, p BuildParams) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")

	cmd.Dir = p.ProjectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func clean(p BuildParams) error {
	bins, err := filepath.Glob(path.Join(p.ProjectPath, "build", "*.out"))
	if err != nil {
		return cerrors.New(err, "failed to fine build files", nil)
	}

	for _, bin := range bins {
		err = os.Remove(bin)
		if err != nil {
			return cerrors.New(err, "failed to delete file", map[string]interface{}{
				"path": bin,
			})
		}
	}

	return nil
}
