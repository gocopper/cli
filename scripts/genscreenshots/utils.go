package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocopper/copper/cerrors"
)

func runCmd(wd, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = wd

	return cmd.Run()
}

func startCmd(wd, name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = wd

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func setCHTTPPort(projectDir string, port int) error {
	err := replaceTextInFile(path.Join(projectDir, "config/base.toml"),
		"port=5901",
		fmt.Sprintf("port=%d", port),
	)
	if err != nil {
		return cerrors.New(err, "failed to update config/base.toml", nil)
	}

	return nil
}

func setVitePort(projectDir string, port int) error {
	err := replaceTextInFile(path.Join(projectDir, "config/dev.toml"),
		"[vitejs]",
		fmt.Sprintf("[vitejs]\nhost=\"http://localhost:%d\"\n", port),
	)
	if err != nil {
		return cerrors.New(err, "failed to update config/dev.toml", nil)
	}

	err = replaceTextInFile(path.Join(projectDir, "web/vite.config.js"),
		"port: 3000",
		fmt.Sprintf("port: %d", port),
	)
	if err != nil {
		return cerrors.New(err, "failed to update web/vite.config.js", nil)
	}

	return nil
}

func replaceTextInFile(path, old, new string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return cerrors.New(err, "failed to open file", map[string]interface{}{
			"path": path,
		})
	}
	defer func() { _ = file.Close() }()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return cerrors.New(err, "failed to read file", map[string]interface{}{
			"path": path,
		})
	}

	updatedData := strings.ReplaceAll(string(data), old, new)

	err = file.Truncate(0)
	if err != nil {
		return cerrors.New(err, "failed to truncate file", map[string]interface{}{
			"path": path,
		})
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return cerrors.New(err, "failed to seek file to 0", map[string]interface{}{
			"path": path,
		})
	}

	_, err = file.WriteString(updatedData)
	if err != nil {
		return cerrors.New(err, "failed to write to file", map[string]interface{}{
			"path": path,
			"data": updatedData,
		})
	}

	return nil
}

func saveScreenshot(url, out string) error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.FullScreenshot(&buf, 90),
	})
	if err != nil {
		return cerrors.New(err, "failed to navigate & take screenshot", map[string]interface{}{
			"url": url,
		})
	}

	err = os.WriteFile(out, buf, 0600)
	if err != nil {
		return cerrors.New(err, "failed to save screenshot file", map[string]interface{}{
			"out": out,
		})
	}

	return nil
}
