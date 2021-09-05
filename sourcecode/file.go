package sourcecode

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gocopper/copper/cerrors"
)

func insertText(path, text string, offset int) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file; %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file; %v", err)
	}

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file; %v", err)
	}

	beg := data[:offset]
	end := data[offset:]

	newData := make([]byte, 0)
	newData = append(newData, beg...)
	newData = append(newData, []byte(text)...)
	newData = append(newData, end...)

	_, err = file.WriteAt(newData, 0)
	if err != nil {
		return fmt.Errorf("failed to write to file; %v", err)
	}

	return nil
}

func CreateTemplateFile(path string, t *template.Template, data interface{}) error {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil && os.IsExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func AppendTemplateToFile(t *template.Template, data interface{}, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s; %v", filePath, err)
	}
	defer file.Close()

	err = t.Execute(file, data)
	if err != nil {
		return cerrors.New(err, "failed to execute template", map[string]interface{}{
			"data": data,
		})
	}

	return nil
}

func InsertTextToFile(path, text string, offset int) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file; %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file; %v", err)
	}

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file; %v", err)
	}

	beg := data[:offset]
	end := data[offset:]

	newData := make([]byte, 0)
	newData = append(newData, beg...)
	newData = append(newData, []byte(text)...)
	newData = append(newData, end...)

	_, err = file.WriteAt(newData, 0)
	if err != nil {
		return fmt.Errorf("failed to write to file; %v", err)
	}

	return nil
}