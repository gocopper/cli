package codemod

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
)

func New(wd string) *CodeModder {
	return &CodeModder{
		wd:   wd,
		data: make(map[string]string),
	}
}

type CodeModder struct {
	wd   string
	data map[string]string
	err  error
}

func (c *CodeModder) CreateTemplateFiles(templatesFS fs.FS, data map[string]string, overwrite bool) *CodeModder {
	if c.err != nil {
		return c
	}

	_, err := createTemplateFiles(templatesFS, c.wd, mergeMaps(c.data, data), overwrite)
	if err != nil {
		return &CodeModder{
			err: err,
		}
	}

	return c
}

func (c *CodeModder) Cd(wd string) *CodeModder {
	if c.err != nil {
		return c
	}

	return New(path.Join(c.wd, wd))
}

func (c *CodeModder) CdAbs(wd string) *CodeModder {
	if c.err != nil {
		return c
	}

	return New(wd)
}

func (c *CodeModder) OpenFile(fp string) *FileModifier {
	if c.err != nil {
		return &FileModifier{
			err: c.err,
		}
	}

	return newFileModifier(path.Join(c.wd, fp), c)
}

func (c *CodeModder) RenameFile(old, new string) *CodeModder {
	if c.err != nil {
		return c
	}

	c.err = os.Rename(path.Join(c.wd, old), path.Join(c.wd, new))

	return c
}

func (c *CodeModder) Remove(fp string) *CodeModder {
	if c.err != nil {
		return c
	}

	c.err = os.Remove(path.Join(c.wd, fp))

	return c
}

func (c *CodeModder) RunCmd(name string, arg ...string) *CodeModder {
	if c.err != nil {
		return c
	}

	cmd := exec.Command(name, arg...)
	cmd.Dir = c.wd
	c.err = cmd.Run()

	return c
}

func (c *CodeModder) ExtractData(fn DataExtractorFn) *CodeModder {
	if c.err != nil {
		return c
	}

	d, err := fn(c.wd)
	if err != nil {
		return &CodeModder{
			err: err,
		}
	}

	c.data = mergeMaps(c.data, d)

	return c
}

func (c *CodeModder) ModifyData(fn func(data map[string]string)) *CodeModder {
	if c.err != nil {
		return c
	}

	fn(c.data)

	return c
}

func (c *CodeModder) Do(fn func(data map[string]string) error) *CodeModder {
	if c.err != nil {
		return c
	}

	c.err = fn(c.data)

	return c
}

func (c *CodeModder) Done() error {
	return c.err
}
