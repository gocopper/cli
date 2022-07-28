package codemod

import (
	"path"
)

func OpenDir(wd string) *Dir {
	return &Dir{
		wd:   wd,
		data: make(map[string]string),
	}
}

type Dir struct {
	wd   string
	data map[string]string
	err  error
}

func (d *Dir) Cd(wd string) *Dir {
	if d.err != nil {
		return d
	}

	return OpenDir(path.Join(d.wd, wd))
}

func (d *Dir) CdAbs(wd string) *Dir {
	if d.err != nil {
		return d
	}

	return OpenDir(wd)
}

func (d *Dir) OpenFile(fp string) *File {
	if d.err != nil {
		return &File{
			err: d.err,
		}
	}

	return openFileInDir(path.Join(d.wd, fp), d)
}

func (d *Dir) ExtractData(fn DataExtractorFn) *Dir {
	if d.err != nil {
		return d
	}

	out, err := fn(d.wd)
	if err != nil {
		return &Dir{
			err: err,
		}
	}

	d.data = mergeMaps(d.data, out)

	return d
}

func (d *Dir) ModifyData(fn func(data map[string]string)) *Dir {
	if d.err != nil {
		return d
	}

	fn(d.data)

	return d
}

func (d *Dir) Apply(fn ...DirFn) *Dir {
	var ret = d
	for i := range fn {
		ret = d.apply(fn[i])
	}
	return ret
}

func (d *Dir) Done() error {
	return d.err
}

func (d *Dir) apply(fn DirFn) *Dir {
	if d.err != nil {
		return d
	}

	err := fn(d.wd, d.data)
	if err != nil {
		return &Dir{err: err}
	}

	return d
}
