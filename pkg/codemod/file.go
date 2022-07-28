package codemod

import (
	"os"

	"github.com/gocopper/copper/cerrors"
)

func openFileInDir(fp string, cm *Dir) *File {
	contents, err := os.ReadFile(fp)
	if err != nil {
		return &File{
			err: cerrors.New(err, "failed to read file", map[string]interface{}{
				"path": fp,
			}),
		}
	}

	return &File{
		path:     fp,
		contents: string(contents),
		err:      nil,
		cm:       cm,
	}
}

type File struct {
	path     string
	contents string
	err      error
	cm       *Dir
}

func (f *File) Apply(fn ...FileFn) *File {
	var ret = f
	for i := range fn {
		ret = f.apply(fn[i])
	}
	return ret
}

func (f *File) Close() *Dir {
	if f.err != nil {
		return &Dir{
			err: f.err,
		}
	}

	err := os.WriteFile(f.path, []byte(f.contents), 0644)
	if err != nil {
		return &Dir{
			err: f.err,
		}
	}

	return f.cm
}

func (f *File) CloseAndOpen(fp string) *File {
	return f.Close().OpenFile(fp)
}

func (f *File) CloseAndDone() error {
	return f.Close().Done()
}

func (f *File) apply(fn FileFn) *File {
	if f.err != nil {
		return f
	}

	d, err := fn(f.contents, f.cm.data)
	if err != nil {
		return &File{
			err: cerrors.New(err, "failed to apply modifier", nil),
		}
	}

	f.contents = d

	return f
}
