package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gocopper/cli/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

type Stack struct {
	Name    string
	RunNPM  bool
	HasVite bool
}

type ScreenGrabber struct {
	Stack             Stack
	CLIPkgPath        string
	HTTPPort          int
	VitePort          int
	ScreenshotsOutDir string
}

func runTask(s *ScreenGrabber) error {
	wd, err := os.MkdirTemp("", "copper-tests-genscreenshots-*")
	if err != nil {
		return cerrors.New(err, "failed to make temp dir", nil)
	}
	defer func() { _ = os.RemoveAll(wd) }()

	var (
		indexURL                  = fmt.Sprintf("http://localhost:%d", s.HTTPPort)
		rocketsURL                = fmt.Sprintf("http://localhost:%d/rockets", s.HTTPPort)
		copperBin                 = path.Join(wd, "copper.out")
		projectDir                = path.Join(wd, "starship")
		screenshotIndexFilePath   = path.Join(s.ScreenshotsOutDir, fmt.Sprintf("%s_index.png", s.Stack.Name))
		screenshotRocketsFilePath = path.Join(s.ScreenshotsOutDir, fmt.Sprintf("%s_rockets.png", s.Stack.Name))
	)

	err = codemod.
		New(s.CLIPkgPath).
		RunCmd("go", "build", "-o", copperBin, ".").
		CdAbs(wd).
		RunCmd(copperBin, "create", "-frontend", s.Stack.Name, "github.com/gocopper/starship").
		CdAbs(projectDir).
		RunCmd(copperBin, "scaffold:pkg", "rockets").
		RunCmd(copperBin, "scaffold:queries", "rockets").
		RunCmd(copperBin, "scaffold:router", "rockets").
		RunCmd(copperBin, "scaffold:route", "-handler", "HandleListRockets", "-path", "/rockets", "rockets").
		OpenFile("./migrations/0001_initial.sql").
		Apply(
			codemod.ModInsertLineAfter(
				"-- +migrate Up",
				`create table rockets (name text); insert into rockets values ('falcon'), ('saturn'), ('atlas');`,
			),
		).
		CloseAndOpen("./pkg/rockets/models.go").
		Apply(
			codemod.ModAppendText(`
type Rocket struct {
	Name string
}`),
		).
		CloseAndOpen("./pkg/rockets/queries.go").
		Apply(
			codemod.ModAddGoImports([]string{"context"}),
			codemod.ModAppendText(`
func (q *Queries) ListRockets(ctx context.Context) ([]Rocket, error) {
	const query = "SELECT * FROM rockets"

	var (
	    rockets []Rocket
	    err = q.querier.Select(ctx, &rockets, query)
    )

	return rockets, err
}`),
		).
		CloseAndOpen("./pkg/rockets/router.go").
		Apply(
			codemod.ModInsertLineAfter("type NewRouterParams struct {", "Queries *Queries"),
			codemod.ModInsertLineAfter("return &Router{", "queries: p.Queries,"),
			codemod.ModInsertLineAfter("type Router struct {", "queries *Queries"),
			codemod.ModInsertLineAfter("HandleListRockets(w http.ResponseWriter, r *http.Request) {", `
	rockets, err := ro.queries.ListRockets(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ro.rw.WriteJSON(w, chttp.WriteJSONParams{
		Data: rockets,
	})
`),
		).
		CloseAndDone()
	if err != nil {
		return cerrors.New(err, "failed to create copper project", nil)
	}

	err = setCHTTPPort(projectDir, s.HTTPPort)
	if err != nil {
		return cerrors.New(err, "failed to set chttp port", map[string]interface{}{
			"dir":  projectDir,
			"port": s.HTTPPort,
		})
	}
	if s.Stack.HasVite {
		err = setVitePort(projectDir, s.VitePort)
		if err != nil {
			return cerrors.New(err, "failed to set vite port", map[string]interface{}{
				"dir":  projectDir,
				"port": s.VitePort,
			})
		}
	}

	cmd, err := startCmd(projectDir, "copper", "run")
	if err != nil {
		return cerrors.New(err, "failed to run copper project", map[string]interface{}{
			"dir": projectDir,
		})
	}
	defer func() { _ = killCmd(cmd) }()

	if s.Stack.RunNPM {
		npmCmd, err := startCmd(path.Join(projectDir, "web"), "npm", "run", "dev")
		if err != nil {
			return cerrors.New(err, "failed to run npm", map[string]interface{}{
				"dir": projectDir,
			})
		}
		defer func() { _ = killCmd(npmCmd) }()
	}

	time.Sleep(15 * time.Second)

	err = saveScreenshot(indexURL, screenshotIndexFilePath)
	if err != nil {
		return cerrors.New(err, "failed to save screenshot", map[string]interface{}{
			"url":      indexURL,
			"filePath": screenshotIndexFilePath,
		})
	}

	err = saveScreenshot(rocketsURL, screenshotRocketsFilePath)
	if err != nil {
		return cerrors.New(err, "failed to save screenshot", map[string]interface{}{
			"url":      rocketsURL,
			"filePath": screenshotRocketsFilePath,
		})
	}

	return nil
}
