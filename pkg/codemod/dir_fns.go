package codemod

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
)

type DirFn func(wd string, data map[string]string) error

func CreateTemplateFiles(templatesFS fs.FS, data map[string]string, overwrite bool) DirFn {
	return func(wd string, additionalData map[string]string) error {
		_, err := createTemplateFiles(templatesFS, wd, mergeMaps(data, additionalData), overwrite)
		return err
	}
}

func RenameFile(old, new string) DirFn {
	return func(wd string, _ map[string]string) error {
		return os.Rename(path.Join(wd, old), path.Join(wd, new))
	}
}

func RemoveFile(fp string) DirFn {
	return func(wd string, _ map[string]string) error {
		return os.Remove(path.Join(wd, fp))
	}
}

func RunCmd(name string, arg ...string) DirFn {
	return func(wd string, data map[string]string) error {
		cmd := exec.Command(name, arg...)
		cmd.Dir = wd

		return cmd.Run()
	}
}
