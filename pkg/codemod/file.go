package codemod

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/gocopper/copper/cerrors"
)

func InsertLineAfterInFile(path, find, text string) error {
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

	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to 0; %v", err)
	}

	_, err = file.WriteString(strings.Replace(string(data), find+"\n", find+"\n"+text+"\n", 1))
	if err != nil {
		return fmt.Errorf("failed to write to file; %v", err)
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

func InsertTemplateToFile(path string, t *template.Template, data interface{}, offset int) error {
	var text strings.Builder

	err := t.Execute(&text, data)
	if err != nil {
		return cerrors.New(err, "failed to execute template", map[string]interface{}{
			"data": data,
		})
	}

	err = InsertTextToFile(path, text.String(), offset)
	if err != nil {
		return cerrors.New(err, "failed to insert text to file", map[string]interface{}{
			"path":   path,
			"offset": offset,
		})
	}

	return nil
}

func AppendTextToFile(path, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s; %v", path, err)
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		return cerrors.New(err, "failed to write text", map[string]interface{}{
			"path": path,
		})
	}

	return nil
}

func AppendTemplateToFile(path string, t *template.Template, data interface{}) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s; %v", path, err)
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
