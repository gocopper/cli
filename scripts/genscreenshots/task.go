package main

import (
	"fmt"
	"os"
	"path"
	"time"

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
		indexURL           = fmt.Sprintf("http://localhost:%d", s.HTTPPort)
		copperBin          = path.Join(wd, "copper.out")
		projectDir         = path.Join(wd, "starship")
		screenshotFilePath = path.Join(s.ScreenshotsOutDir, fmt.Sprintf("%s.png", s.Stack.Name))
	)

	err = runCmd(s.CLIPkgPath, "go", "build", "-o", copperBin, ".")
	if err != nil {
		return cerrors.New(err, "failed to build copper cli", map[string]interface{}{
			"pkgPath": s.CLIPkgPath,
			"out":     copperBin,
		})
	}

	err = runCmd(wd, copperBin, "create", "-frontend", s.Stack.Name, "github.com/gocopper/starship")
	if err != nil {
		return cerrors.New(err, "failed to create copper project", map[string]interface{}{
			"wd":    wd,
			"bin":   copperBin,
			"stack": s.Stack,
		})
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

	err = saveScreenshot(indexURL, screenshotFilePath)
	if err != nil {
		return cerrors.New(err, "failed to save screenshot", map[string]interface{}{
			"url":      indexURL,
			"filePath": screenshotFilePath,
		})
	}

	return nil
}
