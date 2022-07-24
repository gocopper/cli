package codemod

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gocopper/copper/cerrors"
)

func mergeMaps(maps ...map[string]string) map[string]string {
	var ret = make(map[string]string)
	for i := range maps {
		for k, v := range maps[i] {
			ret[k] = v
		}
	}
	return ret
}

func createTemplateFiles(templatesFS fs.FS, dest string, params map[string]string, overwrite bool) ([]string, error) {
	var created = make([]string, 0)

	tmpl, err := template.ParseFS(templatesFS, "tmpl/*.tmpl")
	if err != nil {
		return nil, cerrors.New(err, "failed to parse template files", nil)
	}

	for _, t := range tmpl.Templates() {
		var fpStr strings.Builder

		filePathTmpl, err := template.New(t.Name()).Parse(t.Name())
		if err != nil {
			return nil, cerrors.New(err, "failed to parse file path template", map[string]interface{}{
				"filePath": t.Name(),
			})
		}

		err = filePathTmpl.Execute(&fpStr, params)
		if err != nil {
			return nil, cerrors.New(err, "failed to execute file path template", map[string]interface{}{
				"filePath": t.Name(),
				"params":   params,
			})
		}

		filePath := fpStr.String()
		filePath = strings.ReplaceAll(filePath, "$", "/")
		filePath = strings.Replace(filePath, ".tmpl", "", 1)
		filePath = strings.Replace(filePath, "^", ".", 1)
		filePath = path.Join(dest, filePath)

		err = os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			return created, cerrors.New(err, "failed to create directories for file", map[string]interface{}{
				"file": filePath,
			})
		}

		err = createTemplateFile(t, filePath, params, overwrite)
		if err != nil {
			return created, cerrors.New(err, "failed to create file", map[string]interface{}{
				"file":      filePath,
				"overwrite": overwrite,
			})
		}

		created = append(created, filePath)
	}

	return created, nil
}

func createTemplateFile(t *template.Template, dest string, data interface{}, overwrite bool) error {
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}

	fileFlags := os.O_WRONLY | os.O_CREATE
	if !overwrite {
		fileFlags = fileFlags | os.O_EXCL
	}

	file, err := os.OpenFile(dest, fileFlags, 0644)
	if err != nil && os.IsExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	if overwrite {
		err = file.Truncate(0)
		if err != nil {
			return cerrors.New(err, "failed to truncate file", map[string]interface{}{
				"path": dest,
			})
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return cerrors.New(err, "failed to seek to 0", map[string]interface{}{
				"path": dest,
			})
		}
	}

	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}
